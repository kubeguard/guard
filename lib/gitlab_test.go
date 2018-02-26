package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/appscode/pat"
	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
	"k8s.io/api/authentication/v1"
)

type gitlabTeamRespFunc func(u *url.URL) (int, string)

// GitLab API docs:
// https://docs.gitlab.com/ce/api/README.html#data-validation-and-error-reporting
//
func gitlabGetErrorMsg(err error) []byte {
	//{{{err.Error()}}}
	errMsg := `{ "message": "{{{` + err.Error() + `}}}" }`
	// fmt.Println(errMsg)
	return []byte(errMsg)
}

func gitlabVerifyAuthorization(r *http.Request) error {
	got := r.Header.Get("PRIVATE-TOKEN")
	if got == "" {
		return fmt.Errorf("Header PRIVATE-TOKEN: expected not empty")
	}
	if got != "secret" {
		return fmt.Errorf("PRIVATE-TOKEN: invalid token")
	}
	return nil
}

func gitlabVerifyPageParameter(values []string) (int, error) {
	if len(values) == 1 {
		p, err := strconv.ParseInt(values[0], 10, 32)
		if err != nil {
			return 0, err
		} else {
			return int(p), nil
		}
	} else {
		return 0, fmt.Errorf("invalid query parameter value: %v", values)
	}
}

func gitlabVerifyTeams(teamList []string, expectedSize int) error {
	if len(teamList) != expectedSize {
		return fmt.Errorf("Expected team size: %v, got %v", expectedSize, len(teamList))
	}
	mapTeamName := map[string]bool{}
	for _, name := range teamList {
		mapTeamName[name] = true
	}
	for i := 1; i <= expectedSize; i++ {
		team := "team" + strconv.Itoa(i)
		if _, ok := mapTeamName[team]; !ok {
			return fmt.Errorf("Team %v is missing", team)
		}
	}
	return nil
}

func gitlabVerifyAuthenticatedReview(review *v1.TokenReview, teamSize int) error {
	if !review.Status.Authenticated {
		return fmt.Errorf("Expected authenticated ture, got false")
	}
	if review.Status.User.Username != "nahid" {
		return fmt.Errorf("Expected username %v, got %v", "nahid", review.Status.User.Username)
	}
	if review.Status.User.UID != "1204" {
		return fmt.Errorf("Expected user id %v, got %v", "1204", review.Status.User.UID)
	}
	err := gitlabVerifyTeams(review.Status.User.Groups, teamSize)
	if err != nil {
		return err
	}
	return nil
}

func gitlabVerifyUnauthenticatedReview(review *v1.TokenReview, expectedErr string) error {
	if review.Status.Authenticated {
		return fmt.Errorf("Expected authenticated false, got true")
	}
	if !strings.Contains(review.Status.Error, expectedErr) {
		return fmt.Errorf("Expected error `%v`, got `%v`", expectedErr, review.Status.Error)
	}
	return nil
}

// return string format
//	[
//		{
//			"name":"team1"
//		}
//	]
// team name format : team[teamNo]
func GitlabGetTeams(size int, startTeamNo int) ([]byte, error) {
	type team struct {
		Name string `json:"name"`
	}
	teamList := []team{}
	for i := 1; i <= size; i++ {
		teamList = append(teamList, team{
			Name: string("team" + strconv.Itoa(startTeamNo)),
		})
		startTeamNo++
	}
	teamsInByte, err := json.MarshalIndent(teamList, "", "  ")
	if err != nil {
		return nil, err
	}
	return teamsInByte, nil
}

func gitlabMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func gitlabGetTeamResp(teamSize int) gitlabTeamRespFunc {
	return func(u *url.URL) (int, string) {
		if pg, ok := u.Query()["page"]; ok {
			pageNo, err := gitlabVerifyPageParameter(pg)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
			}

			if perPg, ok := u.Query()["per_page"]; ok {
				perPage, err := gitlabVerifyPageParameter(perPg)
				if err != nil {
					return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
				}
				totalTeams := teamSize
				startTeamNo := (pageNo-1)*perPage + 1
				resp, err := GitlabGetTeams(gitlabMin(totalTeams-startTeamNo+1, perPage), startTeamNo)
				if err != nil {
					return http.StatusInternalServerError, fmt.Sprintf("List user teams request: failed to produce teams. Reason: %v", err)
				}
				return http.StatusOK, string(resp)
			}

			return http.StatusBadRequest, fmt.Sprint("List user teams request: query parameter per_page not provide")

		} else {
			return http.StatusBadRequest, fmt.Sprint("List user teams request: query parameter page not provide")
		}
	}
}

func gitlabServerSetup(userResp string, userStatusCode int, genTeamResp gitlabTeamRespFunc) *httptest.Server {
	m := pat.New()

	m.Get("/user", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := gitlabVerifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(gitlabGetErrorMsg(err))
			return
		}

		w.WriteHeader(userStatusCode)
		w.Write([]byte(userResp))
	}))

	m.Get("/groups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := gitlabVerifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(gitlabGetErrorMsg(err))
			return
		}

		status, resp := genTeamResp(r.URL)
		w.WriteHeader(status)
		if status != http.StatusOK {
			w.Write(gitlabGetErrorMsg(errors.New(resp)))
			return
		}
		w.Write([]byte(resp))
	}))

	srv := httptest.NewServer(m)
	return srv
}

func gitlabClientSetup(serverUrl, token string) (*GitlabClient, error) {
	g := &GitlabClient{
		Client: gitlab.NewClient(nil, token),
	}
	err := g.Client.SetBaseURL(serverUrl)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func TestGitlab(t *testing.T) {
	// https://docs.gitlab.com/ce/api/users.html
	var userRespnBody = `
{
   "id": 1204,
   "username": "nahid"
}
`
	teamSizes := []int{0, 1, 13, 29, 100, 111, 189}
	for _, teamSize := range teamSizes {
		// PerPage=20
		// authenticated : true
		t.Run("scenario 1", func(t *testing.T) {
			srv := gitlabServerSetup(userRespnBody, http.StatusOK, gitlabGetTeamResp(teamSize))
			defer srv.Close()

			client, err := gitlabClientSetup(srv.URL, "secret")
			if err != nil {
				t.Errorf("Error when creating gitlab client. Reason %v", err)
			} else {
				resp, status := client.checkGitLab()
				if status != http.StatusOK {
					t.Errorf("Expected status code 200, got %v. Reason %v", status, resp.Status.Error)
				}
				err := gitlabVerifyAuthenticatedReview(&resp, teamSize)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}

	// authenticated : false
	// reason : invalid token
	t.Run("scenario 2", func(t *testing.T) {
		teamSize := 1
		srv := gitlabServerSetup(userRespnBody, http.StatusOK, gitlabGetTeamResp(teamSize))
		defer srv.Close()

		client, err := gitlabClientSetup(srv.URL, "badtoken")
		if err != nil {
			t.Errorf("Error when creating gitlab client. Reason %v", err)
		} else {
			resp, status := client.checkGitLab()
			if status != http.StatusUnauthorized {
				t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
			}
			err := gitlabVerifyUnauthenticatedReview(&resp, "{{{PRIVATE-TOKEN: invalid token}}}")
			if err != nil {
				t.Error(err)
			}
		}
	})

	// authenticated : false
	// reason : empty token
	t.Run("scenario 3", func(t *testing.T) {
		teamSize := 1
		srv := gitlabServerSetup(userRespnBody, http.StatusOK, gitlabGetTeamResp(teamSize))
		defer srv.Close()
		client, err := gitlabClientSetup(srv.URL, "")
		if err != nil {
			t.Errorf("Error when creating gitlab client. Reason %v", err)
		} else {
			resp, status := client.checkGitLab()
			if status != http.StatusUnauthorized {
				t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
			}
			err := gitlabVerifyUnauthenticatedReview(&resp, "{{{Header PRIVATE-TOKEN: expected not empty}}}")
			if err != nil {
				t.Error(err)
			}
		}
	})

	//error when getting user
	t.Run("scenario 4", func(t *testing.T) {
		teamSize := 1
		errMsg := "error when getting user"
		srv := gitlabServerSetup(string(gitlabGetErrorMsg(errors.New(errMsg))), http.StatusInternalServerError, gitlabGetTeamResp(teamSize))
		defer srv.Close()

		client, err := gitlabClientSetup(srv.URL, "secret")
		if err != nil {
			t.Errorf("Error when creating gitlab client. Reason %v", err)
		} else {
			resp, status := client.checkGitLab()
			if status != http.StatusUnauthorized {
				t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
			}
			err := gitlabVerifyUnauthenticatedReview(&resp, "{{{"+errMsg+"}}}")
			if err != nil {
				t.Error(err)
			}
		}
	})

	//error when getting user's group
	t.Run("scenario 5", func(t *testing.T) {
		errMsg := "error when getting user's group"
		srv := gitlabServerSetup(userRespnBody, http.StatusOK, func(u *url.URL) (int, string) {
			return http.StatusInternalServerError, errMsg
		})
		defer srv.Close()

		client, err := gitlabClientSetup(srv.URL, "secret")
		if err != nil {
			t.Errorf("Error when creating gitlab client. Reason %v", err)
		} else {
			resp, status := client.checkGitLab()
			if status != http.StatusBadRequest {
				t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
			}
			err := gitlabVerifyUnauthenticatedReview(&resp, "{{{"+errMsg+"}}}")
			if err != nil {
				t.Error(err)
			}
		}
	})

	pages := []int{1, 3, 7, 10, 13}
	for _, pageNo := range pages {
		// error when getting user's group at page=[pageNo]
		t.Run("scenario 6", func(t *testing.T) {
			teamSize := 300
			errMsg := fmt.Sprintf("error when getting user's group at page=%v", pageNo)
			srv := gitlabServerSetup(userRespnBody, http.StatusOK, func(u *url.URL) (int, string) {
				if pg, ok := u.Query()["page"]; ok {
					pgNo, err := gitlabVerifyPageParameter(pg)
					if err != nil {
						return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
					}
					if pgNo < pageNo {
						return gitlabGetTeamResp(teamSize)(u)
					} else {
						return http.StatusInternalServerError, errMsg
					}
				} else {
					return http.StatusBadRequest, fmt.Sprint("List user teams request: query parameter page not provide")
				}
			})
			defer srv.Close()

			client, err := gitlabClientSetup(srv.URL, "secret")
			if err != nil {
				t.Errorf("Error when creating gitlab client. Reason %v", err)
			} else {
				resp, status := client.checkGitLab()
				if status != http.StatusBadRequest {
					t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusBadRequest, status, resp.Status.Error)
				}
				err := gitlabVerifyUnauthenticatedReview(&resp, "{{{"+errMsg+"}}}")
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
