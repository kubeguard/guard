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

package token

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	auth "k8s.io/api/authentication/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func stringArrayToBytes(in []string) []byte {
	return []byte(strings.Join(in, "\n"))
}

func assertUserInfo(t *testing.T, got, want auth.UserInfo) {
	if got.Username != want.Username {
		t.Errorf("Expected username %v, got %v", want.Username, got.Username)
	}
	if got.UID != want.UID {
		t.Errorf("Expected uid %v, got %v", want.UID, got.UID)
	}
	if len(got.Groups) != len(want.Groups) {
		t.Errorf("Expected groups size %v, got %v", len(want.Groups), len(got.Groups))
	}
	groupMap := map[string]bool{}
	for _, g := range got.Groups {
		groupMap[g] = true
	}
	for _, g := range want.Groups {
		if !groupMap[g] {
			t.Errorf("Group %v not found", g)
		}
	}
}

func assertLoadTokenResp(t *testing.T, got, want map[string]auth.UserInfo) {
	if len(got) != len(want) {
		t.Errorf("expected item size %v, got %v", len(want), len(got))
	}
	for token, user := range got {
		if wantedUser, found := want[token]; found {
			assertUserInfo(t, user, wantedUser)
		} else {
			t.Errorf("user not found for token %v", token)
		}
	}
}

func TestLoadTokenFile(t *testing.T) {
	loadTokenTests := []struct {
		tokens        []string
		expectedResp  map[string]auth.UserInfo
		expectedError error
	}{
		{
			[]string{
				`token1,user1,1,"group1,group2"`,
				`token2,user2,2,group1`,
				`token3,user3,3,`,
			},
			map[string]auth.UserInfo{
				"token1": {Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
				"token2": {Username: "user2", UID: "2", Groups: []string{"group1"}},
				"token3": {Username: "user3", UID: "3", Groups: []string{}},
			},
			nil,
		},
		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`token2, user2, 2,group1`,
				`token3, user3, 3,`,
			},
			map[string]auth.UserInfo{
				"token1": {Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
				"token2": {Username: "user2", UID: "2", Groups: []string{"group1"}},
				"token3": {Username: "user3", UID: "3", Groups: []string{}},
			},
			nil,
		},
		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`token2, user2, 2,group1`,
				`token3, user3, 3`,
				`token4, user4, 4,`,
			},
			map[string]auth.UserInfo{
				"token1": {Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
				"token2": {Username: "user2", UID: "2", Groups: []string{"group1"}},
				"token3": {Username: "user3", UID: "3"},
				"token4": {Username: "user4", UID: "4"},
			},
			nil,
		},
		{
			[]string{
				`token1,user1,1`,
				`token2, user2, 2`,
				`token3, user3, 3`,
				`token4, user4, 4`,
			},
			map[string]auth.UserInfo{
				"token1": {Username: "user1", UID: "1"},
				"token2": {Username: "user2", UID: "2"},
				"token3": {Username: "user3", UID: "3"},
				"token4": {Username: "user4", UID: "4"},
			},
			nil,
		},
		{
			[]string{
				`token1,user1`,
				`token2, user2`,
				`token3, user3`,
				`token4, user4`,
			},
			nil,
			fmt.Errorf("line #%d of token auth file is ill formatted", 1),
		},
		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`, user2, 2,group1`,
				`token3, user3, 3,"group1"`,
				`token4, user4, 4,"group1"`,
			},
			nil,
			fmt.Errorf("line #%d of token auth file has empty token", 2),
		},
		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`token4, user2, 2,group1`,
				`token3, user3, 3,"group1"`,
				`token4, user4, 4,"group1"`,
			},
			nil,
			fmt.Errorf("line #%d of token auth file reuses token", 4),
		},

		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`token2, user2, 2,group1`,
				`token3, , 3,"group1"`,
			},
			nil,
			fmt.Errorf("line #%d of token auth file has empty user name", 3),
		},
		{
			[]string{
				`token1,user1,1," group1 , group2 "`,
				`token2, user2, 2,group1`,
				`token3, user3, ,"group1"`,
			},
			nil,
			fmt.Errorf("line #%d of token auth file has empty uid", 3),
		},
		{
			[]string{
				`token1,user1,1`,
				`token2, user2, 2`,
				`token3, user1, 1`,
				`token4, user2, 2`,
			},
			map[string]auth.UserInfo{
				"token1": {Username: "user1", UID: "1"},
				"token2": {Username: "user2", UID: "2"},
				"token3": {Username: "user1", UID: "1"},
				"token4": {Username: "user2", UID: "2"},
			},
			nil,
		},
	}

	appFs := afero.NewOsFs()
	filePath := "token-auth/load-file/test"
	err := appFs.MkdirAll(filePath, 0o775)
	if err != nil {
		t.Errorf("Error when making directory. reason : %v", err)
	}
	defer func() {
		utilruntime.Must(appFs.RemoveAll("token-auth"))
	}()

	for _, testData := range loadTokenTests {
		t.Run(fmt.Sprintf("testing load token file, error %v", testData.expectedError), func(t *testing.T) {
			file := filePath + "/token.csv"
			tokenData := stringArrayToBytes(testData.tokens)
			err := afero.WriteFile(appFs, file, tokenData, 0o644)
			if err != nil {
				t.Errorf("Error when creating file. reason : %v", err)
			} else {
				t.Log("test data:", testData)
				resp, err := LoadTokenFile(file)
				if testData.expectedError != nil {
					assert.NotNil(t, err)
					assert.EqualError(t, err, testData.expectedError.Error())
					assert.Nil(t, resp)
				} else {
					assert.Nil(t, err)
					assertLoadTokenResp(t, resp, testData.expectedResp)
				}
			}
		})
	}
}

func TestCheckTokenAuth(t *testing.T) {
	tokenMap := map[string]auth.UserInfo{
		"token1": {Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
		"token2": {Username: "user2", UID: "2", Groups: []string{"group1"}},
		"token3": {Username: "user3", UID: "3", Groups: []string{}},
		"token4": {Username: "user2", UID: "2", Groups: []string{"group2", "group3"}},
	}

	dataset := []struct {
		testName      string
		token         string
		expectedUser  auth.UserInfo
		expectedError string
		authenticated bool
		expectedAuth  bool
	}{
		{
			"authentication successful, multiple groups",
			"token1",
			auth.UserInfo{Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
			"",
			true,
			true,
		},
		{
			"authentication successful, one group",
			"token2",
			auth.UserInfo{Username: "user2", UID: "2", Groups: []string{"group1"}},
			"",
			true,
			true,
		},
		{
			"authentication successful, empty group",
			"token3",
			auth.UserInfo{Username: "user3", UID: "3", Groups: []string{}},
			"",
			true,
			true,
		},
		{
			"authentication successful, same user containing multiple token",
			"token4",
			auth.UserInfo{Username: "user2", UID: "2", Groups: []string{"group2", "group3"}},
			"",
			true,
			true,
		},
		{
			"authentication unsuccessful, reason invalid token",
			"badtoken",
			auth.UserInfo{},
			"Invalid token",
			false,
			false,
		},
		{
			"authentication unsuccessful, reason empty token",
			"",
			auth.UserInfo{},
			"Invalid token",
			false,
			false,
		},
	}

	srv := Authenticator{
		tokenMap: tokenMap,
	}

	for _, testData := range dataset {
		t.Run(testData.testName, func(t *testing.T) {
			resp, err := srv.Check(testData.token)

			t.Log(srv.tokenMap)
			t.Log("token :", testData.token)

			if testData.authenticated {
				assert.Nil(t, err)
				if assert.NotNil(t, resp) {
					assertUserInfo(t, *resp, testData.expectedUser)
				}
			} else {
				assert.NotNil(t, err)
				assert.EqualError(t, err, testData.expectedError)
				assert.Nil(t, resp)
			}
		})
	}
}
