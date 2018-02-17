package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/appscode/guard/lib/graph"
)

type credential struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	TenantID     string `json:"tenantID"`
}

var (
	id_token string = ""
)

func TestGraph(t *testing.T) {
	cred := tGetCred()
	client, err := graph.New(cred.ClientID, cred.ClientSecret, cred.TenantID)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(client.GetGroups("dipta@appscode.com"))

}

func TestCheckAzure(t *testing.T) {
	cred := tGetCred()
	opts := AzureOptions{
		ClientID:     cred.ClientID,
		ClientSecret: cred.ClientSecret,
		TenantID:     cred.TenantID,
	}
	s := Server{Azure: opts}
	resp, status := s.checkAzure(id_token)
	if status != 200 {
		t.Error(resp.Status.Error)
	}
	fmt.Println(resp)
	fmt.Println(resp.Status.User.Username)
	fmt.Println(resp.Status.User.Groups)
}

func tGetCred() credential {
	b, _ := tReadCredFile("/home/ac/Downloads/cred/azure-my.json")
	v := credential{}
	fmt.Println("token-error:", json.Unmarshal(b, &v))
	//fmt.Println(v)
	return v
}

func tReadCredFile(name string) ([]byte, error) {
	crtBytes, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, fmt.Errorf("failed to read `%s`.Reason: %v", name, err)
	}
	return crtBytes, nil
}
