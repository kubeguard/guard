package lib

import (
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

const (
	gitlabUsername     = "nahid"
	gitlabUID          = "1204"
	gitlabGoodToken    = "secret"
	gitlabBadToken     = "badtoken"
	gitlabEmptyToken   = ""
	gitlabUserRespBody = `{ "id": 1204, "username": "nahid" }`
)

type gitlabGroupRespFunc func(u *url.URL) (int, string)

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
	if got != gitlabGoodToken {
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

func gitlabVerifyGroup(groupList []string, expectedSize int) error {
	if len(groupList) != expectedSize {
		return fmt.Errorf("expected group size: %v, got %v", expectedSize, len(groupList))
	}
	mapGroupName := map[string]bool{}
	for _, name := range groupList {
		mapGroupName[name] = true
	}
	for i := 1; i <= expectedSize; i++ {
		group := "team" + strconv.Itoa(i)
		if _, ok := mapGroupName[group]; !ok {
			return fmt.Errorf("group %v is missing", group)
		}
	}
	return nil
}

func gitlabVerifyAuthenticatedReview(review *v1.TokenReview, groupSize int) error {
	if !review.Status.Authenticated {
		return fmt.Errorf("expected authenticated ture, got false")
	}
	if review.Status.User.Username != gitlabUsername {
		return fmt.Errorf("expected username %v, got %v", "nahid", review.Status.User.Username)
	}
	if review.Status.User.UID != gitlabUID {
		return fmt.Errorf("expected user id %v, got %v", "1204", review.Status.User.UID)
	}
	err := gitlabVerifyGroup(review.Status.User.Groups, groupSize)
	if err != nil {
		return err
	}
	return nil
}

func gitlabVerifyUnauthenticatedReview(review *v1.TokenReview, expectedErr string) error {
	if review.Status.Authenticated {
		return fmt.Errorf("expected authenticated false, got true")
	}
	if !strings.Contains(review.Status.Error, expectedErr) {
		return fmt.Errorf("expected error `%v`, got `%v`", expectedErr, review.Status.Error)
	}
	return nil
}

// return string format
//  [
//      {
//          "name":"team1"
//      }
//  ]
// Group name format : team[groupNo]
func GitlabGetGroups(size int, startgroupNo int) ([]byte, error) {
	type group struct {
		Name string `json:"name"`
	}
	groupList := []group{}
	for i := 1; i <= size; i++ {
		groupList = append(groupList, group{
			Name: string("team" + strconv.Itoa(startgroupNo)),
		})
		startgroupNo++
	}
	groupsInByte, err := json.MarshalIndent(groupList, "", "  ")
	if err != nil {
		return nil, err
	}
	return groupsInByte, nil
}

func gitlabMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func gitlabGetGroupResp(groupSize int) gitlabGroupRespFunc {
	return func(u *url.URL) (int, string) {
		if pg, ok := u.Query()["page"]; ok {
			pageNo, err := gitlabVerifyPageParameter(pg)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("List user groups request: %v", err)
			}

			if perPg, ok := u.Query()["per_page"]; ok {
				perPage, err := gitlabVerifyPageParameter(perPg)
				if err != nil {
					return http.StatusBadRequest, fmt.Sprintf("List user groups request: %v", err)
				}
				totalGroups := groupSize
				startGroupNo := (pageNo-1)*perPage + 1
				resp, err := GitlabGetGroups(gitlabMin(totalGroups-startGroupNo+1, perPage), startGroupNo)
				if err != nil {
					return http.StatusInternalServerError, fmt.Sprintf("List user groups request: failed to produce groups. Reason: %v", err)
				}
				return http.StatusOK, string(resp)
			}

			return http.StatusBadRequest, fmt.Sprint("List user groups request: query parameter per_page not provide")

		} else {
			return http.StatusBadRequest, fmt.Sprint("List user groups request: query parameter page not provide")
		}
	}
}

func gitlabServerSetup(userResp string, userStatusCode int, gengroupResp gitlabGroupRespFunc) *httptest.Server {
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

		status, resp := gengroupResp(r.URL)
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

	dataset := []struct {
		testName       string
		userResp       string
		userStatusCode int
		token          string
		expectedErr    string
	}{
		{
			"authentication unsuccessful, reason invalid token",
			gitlabUserRespBody,
			http.StatusOK,
			gitlabBadToken,
			"{{{PRIVATE-TOKEN: invalid token}}}",
		},
		{
			"authentication unsuccessful, reason emtpy token",
			gitlabUserRespBody,
			http.StatusOK,
			gitlabEmptyToken,
			"{{{Header PRIVATE-TOKEN: expected not empty}}}",
		},
		{
			"error when getting user",
			string(gitlabGetErrorMsg(errors.New("error when getting user"))),
			http.StatusInternalServerError,
			gitlabGoodToken,
			"{{{error when getting user}}}",
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			groupSize := 1
			srv := gitlabServerSetup(test.userResp, test.userStatusCode, gitlabGetGroupResp(groupSize))
			defer srv.Close()

			client, err := gitlabClientSetup(srv.URL, test.token)
			if err != nil {
				t.Errorf("Error when creating gitlab client. Reason %v", err)
			} else {
				resp, status := client.checkGitLab()
				if status != http.StatusUnauthorized {
					t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
				}
				err := gitlabVerifyUnauthenticatedReview(&resp, test.expectedErr)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestForDIfferentGroupSizes(t *testing.T) {
	groupSizes := []int{0, 1, 20, 100}
	for _, groupSize := range groupSizes {
		// PerPage=20
		// authenticated : true
		t.Run(fmt.Sprintf("authentication successful, group size %v", groupSize), func(t *testing.T) {
			srv := gitlabServerSetup(gitlabUserRespBody, http.StatusOK, gitlabGetGroupResp(groupSize))
			defer srv.Close()

			client, err := gitlabClientSetup(srv.URL, gitlabGoodToken)
			if err != nil {
				t.Errorf("Error when creating gitlab client. Reason %v", err)
			} else {
				resp, status := client.checkGitLab()
				if status != http.StatusOK {
					t.Errorf("Expected status code 200, got %v. Reason %v", status, resp.Status.Error)
				}
				err := gitlabVerifyAuthenticatedReview(&resp, groupSize)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGroupListErrorInDifferentPage(t *testing.T) {
	pages := []int{1, 2, 3}
	for _, pageNo := range pages {
		t.Run(fmt.Sprintf("error when getting user's group at page %v", pageNo), func(t *testing.T) {
			groupSize := 55
			errMsg := fmt.Sprintf("error when getting user's group at page=%v", pageNo)
			srv := gitlabServerSetup(gitlabUserRespBody, http.StatusOK, func(u *url.URL) (int, string) {
				if pg, ok := u.Query()["page"]; ok {
					pgNo, err := gitlabVerifyPageParameter(pg)
					if err != nil {
						return http.StatusBadRequest, fmt.Sprintf("List user groups request: %v", err)
					}
					if pgNo < pageNo {
						return gitlabGetGroupResp(groupSize)(u)
					} else {
						return http.StatusInternalServerError, errMsg
					}
				} else {
					return http.StatusBadRequest, fmt.Sprint("List user groups request: query parameter page not provide")
				}
			})
			defer srv.Close()

			client, err := gitlabClientSetup(srv.URL, gitlabGoodToken)
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
