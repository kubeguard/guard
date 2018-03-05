package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/xanzy/go-gitlab"
)

func TestGitLab(t *testing.T) {
	resp, status := checkGitLab(tGetToken())
	if status != http.StatusOK {
		t.Error("Expected", http.StatusOK, ". Got", status)
	}
	fmt.Println(resp)
}

func TestSome(t *testing.T) {
	client := gitlab.NewClient(nil, tGetToken())
	grup, _, err := client.Groups.ListGroups(nil)
	if err != nil {
		t.Error(err)
	}
	for _, g := range grup {
		fmt.Println(g.Name)
	}
}

func tGetToken() string {
	b, _ := tReadFile("/home/ac/Downloads/cred/gitlabR.json")
	v := struct {
		Token string `json:"token"`
	}{}
	fmt.Println("token-error:", json.Unmarshal(b, &v))
	//fmt.Println(v)
	return v.Token
}

func tReadFile(name string) ([]byte, error) {
	crtBytes, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, errors.Errorf("failed to read `%s`.Reason: %v", name, err)
	}
	return crtBytes, nil
}
