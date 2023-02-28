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

package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/appscode/pat"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	gitlabUsername     = "nahid"
	gitlabUID          = "1204"
	gitlabGoodToken    = "secret"
	gitlabBadToken     = "badtoken"
	gitlabEmptyToken   = ""
	gitlabUserRespBody = `{ "id": 1204, "username": "nahid" }`
)

var testGroupID = map[bool]string{
	true:  ":using-group-id",
	false: ":using-group-fullpath",
}

type gitlabGroupRespFunc func(u *url.URL) (int, string)

// GitLab API docs:
// https://docs.gitlab.com/ce/api/README.html#data-validation-and-error-reporting
//
func gitlabGetErrorMsg(err error) []byte {
	// {{{err.Error()}}}
	errMsg := `{ "message": "{{{` + err.Error() + `}}}" }`
	// fmt.Println(errMsg)
	return []byte(errMsg)
}

func gitlabVerifyAuthorization(r *http.Request) error {
	got := r.Header.Get("PRIVATE-TOKEN")
	if got == "" {
		return fmt.Errorf("header PRIVATE-TOKEN: expected not empty")
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

func assertGroup(t *testing.T, useGroupId bool, groupList []string, expectedSize int) {
	if len(groupList) != expectedSize {
		t.Errorf("expected group size: %v, got %v", expectedSize, len(groupList))
	}

	groups := sets.NewString(groupList...)
	for i := 1; i <= expectedSize; i++ {
		group := strconv.Itoa(i)
		if !useGroupId {
			group = "team" + group
		}
		if !groups.Has(group) {
			t.Errorf("group %v is missing", group)
		}
	}
}

func assertUserInfo(t *testing.T, info *v1.UserInfo, useGroupId bool, groupSize int) {
	if info.Username != gitlabUsername {
		t.Errorf("expected username %v, got %v", "nahid", info.Username)
	}
	if info.UID != gitlabUID {
		t.Errorf("expected user id %v, got %v", "1204", info.UID)
	}
	assertGroup(t, useGroupId, info.Groups, groupSize)
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
		ID       int    `json:"id"`
		FullPath string `json:"full_path"`
	}
	var groupList []group
	for i := 1; i <= size; i++ {
		groupList = append(groupList, group{
			ID:       startgroupNo,
			FullPath: string("team" + strconv.Itoa(startgroupNo)),
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

			return http.StatusBadRequest, "List user groups request: query parameter per_page not provide"
		}
		return http.StatusBadRequest, "List user groups request: query parameter page not provide"
	}
}

func gitlabServerSetup(userResp string, userStatusCode int, gengroupResp gitlabGroupRespFunc) *httptest.Server {
	m := pat.New()

	m.Get("/api/v4/user", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := gitlabVerifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(gitlabGetErrorMsg(err))
			return
		}

		w.WriteHeader(userStatusCode)
		_, _ = w.Write([]byte(userResp))
	}))

	m.Get("/api/v4/groups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := gitlabVerifyAuthorization(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write(gitlabGetErrorMsg(err))
			return
		}

		status, resp := gengroupResp(r.URL)
		w.WriteHeader(status)
		if status != http.StatusOK {
			_, _ = w.Write(gitlabGetErrorMsg(errors.New(resp)))
			return
		}
		_, _ = w.Write([]byte(resp))
	}))

	srv := httptest.NewServer(m)
	return srv
}

func gitlabClientSetup(serverUrl string, useGroupId bool) *Authenticator {
	g := &Authenticator{
		opts: Options{
			BaseUrl:    serverUrl,
			UseGroupID: useGroupId,
		},
	}

	return g
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
			"authentication unsuccessful, reason empty token",
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
	ctx := context.Background()

	for _, test := range dataset {
		for useGroupId, suffix := range testGroupID {
			t.Run(test.testName+suffix, func(t *testing.T) {
				groupSize := 1
				srv := gitlabServerSetup(test.userResp, test.userStatusCode, gitlabGetGroupResp(groupSize))
				defer srv.Close()

				client := gitlabClientSetup(srv.URL, useGroupId)

				resp, err := client.Check(ctx, test.token)
				if assert.NotNil(t, err) {
					assert.Nil(t, resp)
				}
			})
		}
	}
}

func TestForDIfferentGroupSizes(t *testing.T) {
	groupSizes := []int{0, 1, 20, 100}
	ctx := context.Background()

	for _, groupSize := range groupSizes {
		// PerPage=20
		// authenticated : true
		for useGroupId, suffix := range testGroupID {
			t.Run(fmt.Sprintf("authentication successful, group size %v %s", groupSize, suffix), func(t *testing.T) {
				srv := gitlabServerSetup(gitlabUserRespBody, http.StatusOK, gitlabGetGroupResp(groupSize))
				defer srv.Close()

				client := gitlabClientSetup(srv.URL, useGroupId)
				if assert.NotNil(t, client) {
					resp, err := client.Check(ctx, gitlabGoodToken)
					if assert.Nil(t, err) {
						assertUserInfo(t, resp, useGroupId, groupSize)
					}
				}
			})
		}
	}
}

func TestGroupListErrorInDifferentPage(t *testing.T) {
	pages := []int{1, 2, 3}
	ctx := context.Background()

	for _, pageNo := range pages {
		for useGroupId, suffix := range testGroupID {
			t.Run(fmt.Sprintf("error when getting user's group at page %v %s", pageNo, suffix), func(t *testing.T) {
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
						return http.StatusBadRequest, "List user groups request: query parameter page not provide"
					}
				})
				defer srv.Close()

				client := gitlabClientSetup(srv.URL, useGroupId)
				resp, err := client.Check(ctx, gitlabGoodToken)
				assert.NotNil(t, err)
				assert.Nil(t, resp)
			})
		}
	}
}
