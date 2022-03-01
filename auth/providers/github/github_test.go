/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/appscode/pat"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	githubOrganization = "appscode"
	githubGoodToken    = "secret"
	githubBadToken     = "badtoken"
	githubUsername     = "nahid"
	githubUID          = "1204"
	githubMemRespBody  = `{ "user":{ "login":"nahid", "id":1204 } }`
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

			return http.StatusBadRequest, "List user teams request: query parameter per_page not provide"
		}
		return http.StatusBadRequest, "List user teams request: query parameter page not provide"
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

func assertTeamList(t *testing.T, teamList []string, expectedSize int) {
	if len(teamList) != expectedSize {
		t.Errorf("Expected team size: %v, got %v", expectedSize, len(teamList))
	}

	teams := sets.NewString(teamList...)
	for i := 1; i <= expectedSize; i++ {
		team := "team" + strconv.Itoa(i)
		if !teams.Has(team) {
			t.Errorf("Team %v is missing", team)
		}
	}
}

func assertUserInfo(t *testing.T, info *v1.UserInfo, teamSize int) {
	if info.Username != githubUsername {
		t.Errorf("Expected username %v, got %v", "nahid", info.Username)
	}
	if info.UID != githubUID {
		t.Errorf("Expected user id %v, got %v", "1204", info.UID)
	}
	assertTeamList(t, info.Groups, teamSize)
}

func githubServerSetup(githubOrg string, memberResp string, memberStatusCode int, genTeamRespn teamRespFunc) *httptest.Server {
	m := pat.New()

	m.Get(fmt.Sprintf("/user/memberships/orgs/%v", githubOrg), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := verifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(getErrorMessage(err))
			return
		}
		w.WriteHeader(memberStatusCode)
		_, _ = w.Write([]byte(memberResp))
	}))

	m.Get("/user/memberships/orgs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(getErrorMessage(errors.New("Authorization: invalid token")))
	}))

	m.Get("/user/teams", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := verifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(getErrorMessage(err))
			return
		}

		status, resp := genTeamRespn(r.URL)
		w.WriteHeader(status)
		if status != http.StatusOK {
			_, _ = w.Write(getErrorMessage(errors.New(resp)))
			return
		}
		_, _ = w.Write([]byte(resp))
	}))

	srv := httptest.NewServer(m)
	return srv
}

func githubClientSetup(serverUrl, githubOrg string) *Authenticator {
	g := &Authenticator{
		opts: Options{
			BaseUrl: serverUrl,
		},
		ctx:     context.Background(),
		OrgName: githubOrg,
	}
	return g
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
			githubMemRespBody,
			http.StatusOK,
			githubOrganization,
			githubOrganization,
			githubBadToken,
			"{{{Authorization: invalid token}}}",
		},
		{
			"authentication unsuccessful, error: invalid token, org used: code",
			githubMemRespBody,
			http.StatusOK,
			githubOrganization,
			"code",
			githubGoodToken,
			"{{{Authorization: invalid token}}}",
		},
		{
			"error when getting user organization membership",
			string(getErrorMessage(errors.New("error when checking organization membership"))),
			http.StatusUnauthorized,
			githubOrganization,
			githubOrganization,
			githubGoodToken,
			"{{{error when checking organization membership}}}",
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			t.Log(test)
			teamSize := 1
			srv := githubServerSetup(test.org, test.memRespBody, test.memStatusCode, getTeamRespFunc(teamSize))
			defer srv.Close()

			client := githubClientSetup(srv.URL, test.reqOrg)

			resp, err := client.Check(test.accessToken)
			assert.NotNil(t, err)
			assert.Nil(t, resp)
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
			srv := githubServerSetup(githubOrganization, githubMemRespBody, http.StatusOK, getTeamRespFunc(teamSize))
			defer srv.Close()

			client := githubClientSetup(srv.URL, githubOrganization)

			resp, err := client.Check(githubGoodToken)
			assert.Nil(t, err)
			assertUserInfo(t, resp, teamSize)
		})
	}
}

func TestAuthorizationHeader(t *testing.T) {
	teamSize := 1
	srv := githubServerSetup(githubOrganization, githubMemRespBody, http.StatusOK, getTeamRespFunc(teamSize))
	defer srv.Close()

	client := githubClientSetup(srv.URL, githubOrganization)

	resp, err := client.Check("")
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestTeamListErrorAtDifferentPage(t *testing.T) {
	pages := []int{1, 2, 3}
	for _, pageNo := range pages {
		// error when getting user's team list at page=[pageNo]
		t.Run(fmt.Sprintf("error when getting user's team list at page %v", pageNo), func(t *testing.T) {
			teamSize := 60 // 3 pages
			errMsg := fmt.Sprintf("error when getting user's team list at page=%v", pageNo)

			srv := githubServerSetup(githubOrganization, githubMemRespBody, http.StatusOK, func(u *url.URL) (int, string) {
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
					return http.StatusBadRequest, "List user teams request: query parameter page not provide"
				}
			})
			defer srv.Close()

			client := githubClientSetup(srv.URL, githubOrganization)

			resp, err := client.Check(githubGoodToken)
			assert.NotNil(t, err)
			assert.Nil(t, resp)
		})
	}
}
