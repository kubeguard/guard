package lib

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/spf13/afero"
	auth "k8s.io/api/authentication/v1"
)

func stringArraytoBytes(in []string) []byte {
	out := ""
	for _, s := range in {
		if len(out) == 0 {
			out = s
		} else {
			out = out + "\n" + s
		}
	}
	return []byte(out)
}

func verifyUserInfo(got, want auth.UserInfo) error {
	if got.Username != want.Username {
		return fmt.Errorf("Expected username %v, got %v", want.Username, got.Username)
	}
	if got.UID != want.UID {
		return fmt.Errorf("Expected uid %v, got %v", want.UID, got.UID)
	}
	if len(got.Groups) != len(want.Groups) {
		return fmt.Errorf("Expected groups size %v, got %v", len(want.Groups), len(got.Groups))
	}
	groupMap := map[string]bool{}
	for _, g := range got.Groups {
		groupMap[g] = true
	}
	for _, g := range want.Groups {
		if !groupMap[g] {
			return fmt.Errorf("Group %v not found", g)
		}
	}
	return nil
}

func verifyLoadTokenResp(got, want map[string]auth.UserInfo) error {
	if len(got) != len(want) {
		return fmt.Errorf("expected item size %v, got %v", len(want), len(got))
	}
	for token, user := range got {
		if wantedUser, found := want[token]; found {
			if err := verifyUserInfo(user, wantedUser); err != nil {
				return fmt.Errorf("Expected user %v, got %v, error : %v", wantedUser, user, err)
			}
		} else {
			return fmt.Errorf("user not found for token %v", token)
		}
	}
	return nil
}

func checkError(got, want error) error {

	if want == nil || got == nil {
		if want != got {
			return fmt.Errorf("Error: expected %v, got %v", want, got)
		}
	} else {
		if want.Error() != got.Error() {
			return fmt.Errorf("Error: expected %v, got %v", want, got)
		}
	}
	return nil
}

func TestLoadTokenFile(t *testing.T) {
	var loadTokenTests = []struct {
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
				`token4, user3, 3`,
				`token3, user4, 4,`,
			},
			nil,
			fmt.Errorf("failed to parse token auth file: line 3, column 0: wrong number of fields in line"),
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
	appFs.MkdirAll(filePath, 0775)
	defer appFs.RemoveAll("token-auth")

	for _, testData := range loadTokenTests {
		t.Run("scenario 1", func(t *testing.T) {
			file := filePath + "/token.csv"
			tokenData := stringArraytoBytes(testData.tokens)
			err := afero.WriteFile(appFs, file, tokenData, 0644)
			if err != nil {
				t.Errorf("Error when creating file. reason : %v", err)
			} else {
				t.Log("test data:", testData)
				resp, err := LoadTokenFile(file)
				if err := checkError(err, testData.expectedError); err != nil {
					t.Log(string(tokenData))
					t.Error(err)
				} else if err := verifyLoadTokenResp(resp, testData.expectedResp); err != nil {
					t.Log(string(tokenData))
					t.Errorf("UserInfo: %v", err)
				}
			}
		})
	}
}

func TestCheckTokenAuth(t *testing.T) {
	previousTokenMap := tokenMap
	tokenMap = map[string]auth.UserInfo{
		"token1": {Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
		"token2": {Username: "user2", UID: "2", Groups: []string{"group1"}},
		"token3": {Username: "user3", UID: "3", Groups: []string{}},
		"token4": {Username: "user2", UID: "2", Groups: []string{"group2", "group3"}},
	}

	dataset := []struct {
		token          string
		expectedUser   auth.UserInfo
		expectedError  string
		expectedStatus int
		expectedAuth   bool
	}{
		{
			"token1",
			auth.UserInfo{Username: "user1", UID: "1", Groups: []string{"group1", "group2"}},
			"",
			http.StatusOK,
			true,
		},
		{
			"token2",
			auth.UserInfo{Username: "user2", UID: "2", Groups: []string{"group1"}},
			"",
			http.StatusOK,
			true,
		},
		{
			"token3",
			auth.UserInfo{Username: "user3", UID: "3", Groups: []string{}},
			"",
			http.StatusOK,
			true,
		},
		{
			"token4",
			auth.UserInfo{Username: "user2", UID: "2", Groups: []string{"group2", "group3"}},
			"",
			http.StatusOK,
			true,
		},
		{
			"badtoken",
			auth.UserInfo{},
			"Invalid token",
			http.StatusUnauthorized,
			false,
		},
		{
			"",
			auth.UserInfo{},
			"Invalid token",
			http.StatusUnauthorized,
			false,
		},
	}

	srv := Server{}

	for pos, testData := range dataset {
		t.Run(fmt.Sprintf("scenario 1 testcase #%v", pos), func(t *testing.T) {
			resp, status := srv.checkTokenAuth(testData.token)

			t.Log(tokenMap)
			t.Log("token :", testData.token)

			if status != testData.expectedStatus {
				t.Errorf("Expected status code %v, got %v", testData.expectedStatus, status)
			}

			if resp.Status.Authenticated != testData.expectedAuth {
				t.Errorf("Expected Authentication %v, got %v", testData.expectedAuth, resp.Status.Authenticated)
			}

			if resp.Status.Authenticated {
				if err := verifyUserInfo(resp.Status.User, testData.expectedUser); err != nil {
					t.Error(err)
					t.Errorf("Expected user %v, got %v", testData.expectedUser, resp.Status.User)
				}
			} else {
				if resp.Status.Error != testData.expectedError {
					t.Errorf("Expected error message %v, got %v", testData.expectedError, resp.Status.Error)
				}
			}
		})
	}

	//restoring tokenMap in it's previous state
	tokenMap = previousTokenMap
}
