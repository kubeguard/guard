package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/appscode/pat"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"k8s.io/api/authentication/v1"
)

const (
	organization = "appscode"
	goodToken    = "secret"
	badToken     = "badtoken"
	username     = "nahid"
	uid          = "1204"
	memRespBody  = `{ "user":{ "login":"nahid", "id":1204 } }`
)

type teamRespFunc func(u *url.URL) (int, string)

// return string format
//	[
//		{
//			"organization":{
//				"login":"appscode"
//			},
//			"name":"team1"
//		}
//	]
// team name format : team[teamNo]
func getTeamList(size int, startTeamNo int) ([]byte, error) {
	type team struct {
		Organization struct {
			Login string `json:"login"`
		} `json:"organization"`
		Name string `json:"name"`
	}
	teamList := []team{}
	for i := 1; i <= size; i++ {
		teamList = append(teamList, team{
			Organization: struct {
				Login string `json:"login"`
			}{
				Login: "appscode",
			},
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

func verifyAuthorization(r *http.Request) error {
	got := r.Header.Get("Authorization")
	if got == "" {
		return fmt.Errorf("Header Authorization: expected not empty")
	}
	if got != "Bearer secret" {
		return fmt.Errorf("Authorization: invalid token")
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other
// response body will be silently ignored.
//
//An ErrorResponse reports one or more errors caused by an API request.
//
//GitHub API docs: https://developer.github.com/v3/#client-errors
//
//type ErrorResponse struct {
//	Response *http.Response // HTTP response that caused this error
//	Message  string         `json:"message"` // error message
//	Errors   []Error        `json:"errors"`  // more detail on individual errors
//	// Block is only populated on certain types of errors such as code 451.
//	// See https://developer.github.com/changes/2016-03-17-the-451-status-code-is-now-supported/
//	// for more information.
//	Block *struct {
//		Reason    string     `json:"reason,omitempty"`
//		CreatedAt *Timestamp `json:"created_at,omitempty"`
//	} `json:"block,omitempty"`
//	// Most errors will also include a documentation_url field pointing
//	// to some content that might help you resolve the error, see
//	// https://developer.github.com/v3/#client-errors
//	DocumentationURL string `json:"documentation_url,omitempty"`
//}
func getErrorMessage(err error) []byte {
	//{{{err.Error()}}}
	errMsg := `{ "message": "{{{` + err.Error() + `}}}" }`
	// fmt.Println(errMsg)
	return []byte(errMsg)
}

func getTeamRespFunc(teamSize int) teamRespFunc {
	return func(u *url.URL) (int, string) {
		if pg, ok := u.Query()["page"]; ok {
			pageNo, err := verifyPageParameter(pg)
			if err != nil {
				return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
			}

			if perPg, ok := u.Query()["per_page"]; ok {
				perPage, err := verifyPageParameter(perPg)
				if err != nil {
					return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
				}
				totalTeams := teamSize
				startTeamNo := (pageNo-1)*perPage + 1
				resp, err := getTeamList(min(totalTeams-startTeamNo+1, perPage), startTeamNo)
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

func verifyPageParameter(values []string) (int, error) {
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

func verifyTeamList(teamList []string, expectedSize int) error {
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

func verifyAuthenticatedTokenReview(review *v1.TokenReview, teamSize int) error {
	if !review.Status.Authenticated {
		return fmt.Errorf("Expected authenticated ture, got false")
	}
	if review.Status.User.Username != username {
		return fmt.Errorf("Expected username %v, got %v", "nahid", review.Status.User.Username)
	}
	if review.Status.User.UID != uid {
		return fmt.Errorf("Expected user id %v, got %v", "1204", review.Status.User.UID)
	}
	err := verifyTeamList(review.Status.User.Groups, teamSize)
	if err != nil {
		return err
	}
	return nil
}

func verifyUnauthenticatedTokenReview(review *v1.TokenReview, expectedErr string) error {
	if review.Status.Authenticated {
		return fmt.Errorf("Expected authenticated false, got true")
	}
	if !strings.Contains(review.Status.Error, expectedErr) {
		return fmt.Errorf("Expected error `%v`, got `%v`", expectedErr, review.Status.Error)
	}
	return nil
}

func githubServerSetup(githubOrg string, memberResp string, memberStatusCode int, genTeamRespn teamRespFunc) *httptest.Server {
	m := pat.New()

	m.Get(fmt.Sprintf("/user/memberships/orgs/%v", githubOrg), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := verifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(getErrorMessage(err))
			return
		}
		w.WriteHeader(memberStatusCode)
		w.Write([]byte(memberResp))
	}))

	m.Get(fmt.Sprintf("/user/memberships/orgs/"), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(getErrorMessage(errors.New("Authorization: invalid token")))
		return
	}))

	m.Get("/user/teams", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := verifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(getErrorMessage(err))
			return
		}

		status, resp := genTeamRespn(r.URL)
		w.WriteHeader(status)
		if status != http.StatusOK {
			w.Write(getErrorMessage(errors.New(resp)))
			return
		}
		w.Write([]byte(resp))
	}))

	srv := httptest.NewServer(m)
	return srv
}

func githubClientSetup(serverUrl, githubOrg string, ctx context.Context, httpClient *http.Client) (*GithubClient, error) {
	g := &GithubClient{
		Ctx:     ctx,
		OrgName: githubOrg,
	}
	var err error
	g.Client, err = github.NewEnterpriseClient(serverUrl, serverUrl, httpClient)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func TestCheckGithub(t *testing.T) {

	dataset := []struct {
		testName      string
		memRespBody   string
		memStatusCode int
		org           string
		reqOrg        string
		accessToken   string
		expectedErr   string
	}{
		{
			"authentication unsuccessful, error: invalid token",
			memRespBody,
			http.StatusOK,
			organization,
			organization,
			badToken,
			"{{{Authorization: invalid token}}}",
		},
		{
			"authentication unsuccessful, error: invalid token, org used: code",
			memRespBody,
			http.StatusOK,
			organization,
			"code",
			goodToken,
			"{{{Authorization: invalid token}}}",
		},
		{
			"error when getting user organization membership",
			string(getErrorMessage(errors.New("error when checking organization membership"))),
			http.StatusUnauthorized,
			organization,
			organization,
			goodToken,
			"{{{error when checking organization membership}}}",
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			t.Log(test)
			teamSize := 1
			srv := githubServerSetup(test.org, test.memRespBody, test.memStatusCode, getTeamRespFunc(teamSize))
			defer srv.Close()

			ctx := context.Background()
			client, err := githubClientSetup(srv.URL, test.reqOrg, ctx, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: test.accessToken},
			)))

			if err != nil {
				t.Errorf("Error when creating github client. Reason %v", err)
			} else {
				resp, status := client.checkGithub()
				if status != http.StatusUnauthorized {
					t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
				}

				err := verifyUnauthenticatedTokenReview(&resp, test.expectedErr)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestForDifferentTeamSizes(t *testing.T) {
	teamSizes := []int{25, 0, 1, 100} // 25 * N

	for _, size := range teamSizes {
		// page=1, PerPage=25
		// authenticated : true
		t.Run(fmt.Sprintf("authentication successful, team size: %v", size), func(t *testing.T) {
			teamSize := size
			srv := githubServerSetup(organization, memRespBody, http.StatusOK, getTeamRespFunc(teamSize))
			defer srv.Close()
			ctx := context.Background()
			client, err := githubClientSetup(srv.URL, organization, ctx, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: goodToken},
			)))
			if err != nil {
				t.Errorf("Error when creating github client. Reason %v", err)
			} else {
				resp, status := client.checkGithub()
				if status != http.StatusOK {
					t.Errorf("Expected status code 200, got %v. Reason %v", status, resp.Status.Error)
				}
				err := verifyAuthenticatedTokenReview(&resp, teamSize)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestAuthorizationHeader(t *testing.T) {
	teamSize := 1
	srv := githubServerSetup(organization, memRespBody, http.StatusOK, getTeamRespFunc(teamSize))
	defer srv.Close()
	ctx := context.Background()
	client, err := githubClientSetup(srv.URL, organization, ctx, nil)
	if err != nil {
		t.Errorf("Error when creating github client. Reason %v", err)
	} else {
		resp, status := client.checkGithub()
		if status != http.StatusUnauthorized {
			t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
		}
		err := verifyUnauthenticatedTokenReview(&resp, "{{{Header Authorization: expected not empty}}}")
		if err != nil {
			t.Error(err)
		}
	}
}

func TestTeamListErrorAtDifferentPage(t *testing.T) {
	pages := []int{1, 2, 3}
	for _, pageNo := range pages {
		// error when getting user's team list at page=[pageNo]
		t.Run(fmt.Sprintf("error when getting user's team list at page %v", pageNo), func(t *testing.T) {
			teamSize := 60
			errMsg := fmt.Sprintf("error when getting user's team list at page=%v", pageNo)

			srv := githubServerSetup(organization, memRespBody, http.StatusOK, func(u *url.URL) (int, string) {
				if pg, ok := u.Query()["page"]; ok {
					pgNo, err := verifyPageParameter(pg)
					if err != nil {
						return http.StatusBadRequest, fmt.Sprintf("List user teams request: %v", err)
					}
					if pgNo < pageNo {
						return getTeamRespFunc(teamSize)(u)
					} else {
						return http.StatusInternalServerError, errMsg
					}
				} else {
					return http.StatusBadRequest, fmt.Sprint("List user teams request: query parameter page not provide")
				}
			})
			defer srv.Close()

			ctx := context.Background()
			client, err := githubClientSetup(srv.URL, organization, ctx, oauth2.NewClient(ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: goodToken},
			)))
			if err != nil {
				t.Errorf("Error when creating github client. Reason %v", err)
			} else {
				resp, status := client.checkGithub()
				if status != http.StatusUnauthorized {
					t.Errorf("Expected status code %v, got %v. Reason %v", http.StatusUnauthorized, status, resp.Status.Error)
				}
				err := verifyUnauthenticatedTokenReview(&resp, "{{{"+errMsg+"}}}")
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
