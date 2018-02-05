package lib

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	auth "k8s.io/api/authentication/v1beta1"
)

var (
	tokenAuthCsvFile string = ""
)

type tokenInfo struct {
	token    string   `json:"token"`
	userName string   `json:"username"`
	userID   string   `json:"userid"`
	groups   []string `json:"groups"`
}

func checkTokenAuth(token string) (auth.TokenReview, int) {
	tokenList, err := ParseCsvFile(tokenAuthCsvFile)
	if err != nil {
		return Error(fmt.Sprintf("Failed to load data from token auth file. Reason: %v", err)), http.StatusInternalServerError
	}

	data := auth.TokenReview{}

	for _, t := range tokenList {
		if t.token == token {
			data.Status = auth.TokenReviewStatus{
				User: auth.UserInfo{
					Username: t.userName,
					UID:      t.userID,
					Groups:   t.groups,
				},
			}
			data.Status.Authenticated = true
			return data, http.StatusOK
		}
	}

	return Error("Invalid token"), http.StatusUnauthorized
}

func ReadCvsFile(file string) ([]byte, error) {
	err := ValidateTokenCvsFile(file)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// https://kubernetes.io/docs/admin/authentication/#static-token-file
func ParseCsvFile(file string) ([]tokenInfo, error) {
	err := ValidateTokenCvsFile(file)
	if err != nil {
		return nil, err
	}
	csvFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	data := []tokenInfo{}
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		info := tokenInfo{
			token:    strings.Trim(line[0], " "),
			userName: strings.Trim(line[1], " "),
			userID:   strings.Trim(line[2], " "),
		}
		if len(line) > 3 {
			info.groups = ParseGroupFromString(strings.Trim(line[3], " "))
		}
		data = append(data, info)
	}
	return data, nil
}

//string format : "group1,group2,group3"
func ParseGroupFromString(in string) []string {
	out := []string{}
	groups := strings.Split(in, ",")
	for _, g := range groups {
		if len(g) > 0 {
			out = append(out, strings.Trim(g, " "))
		}
	}
	return out
}

//https://kubernetes.io/docs/admin/authentication/#static-token-file
//csv token file:
//  - four field required (format : token,user,uid,"group1,group2,group3")
//  - groups can be empty, others cannot be empty
//  - token should be unique
//  - one user can have multiple token
func ValidateTokenCvsFile(file string) error {
	csvFile, err := os.Open(file)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	tokenUsed := map[string]bool{}
	lineCount := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		lineCount++
		if len(line) != 4 {
			return fmt.Errorf("in the token file, line %v has insufficient number of fields", lineCount)
		}
		token := strings.Trim(line[0], " ")
		if len(token) == 0 {
			return fmt.Errorf("token cannot be empty")
		}
		if found := tokenUsed[token]; found {
			return fmt.Errorf("token must be unique")
		}
		tokenUsed[token] = true

		if len(strings.Trim(line[1], " ")) == 0 {
			return fmt.Errorf("user cannot be empty")
		}
		if len(strings.Trim(line[2], " ")) == 0 {
			return fmt.Errorf("uid cannot be empty")
		}
	}
	if lineCount == 0 {
		return fmt.Errorf("empty token file")
	}
	return nil
}
