package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/appscode/pat"
	oidc "github.com/coreos/go-oidc"
	gdir "google.golang.org/api/admin/directory/v1"
	auth "k8s.io/api/authentication/v1"
)

const (
	userEmail   = "nahid@domain.com"
	adminEmail  = "admin@domain.com"
	domain      = "domain"
	googleToken = `{ "iss" : "%v", "email" : "%v", "aud" : "%v", "hd" : "%v"}`
)

type googleGroupResp func(u *url.URL) (int, []byte)

// group email format: group[groupNo]@domain.com
func googleGetGroupEmail(groupNo int) string {
	return "group" + strconv.Itoa(groupNo) + "@domain.com"
}

func numberOfPage(total, perPage int) int {
	x := float64(total)
	y := float64(perPage)
	return int(math.Ceil(x / y))
}

func googleGetGroups(pageToken string, groupStartNo, groupPerPage int) gdir.Groups {
	groups := gdir.Groups{
		NextPageToken: pageToken,
	}
	for i := 1; i <= groupPerPage; i++ {
		groups.Groups = append(groups.Groups, &gdir.Group{
			Email: googleGetGroupEmail(groupStartNo + i - 1),
		})
	}
	return groups
}

func googleGetGroupResp(totalGroups, groupPerPage, totalPage int) googleGroupResp {

	groupData := map[string]gdir.Groups{}
	for i := 1; i <= totalPage; i++ {
		nextPageToken := strconv.Itoa(i + 1)
		if i == totalPage {
			// empty next page token for last page
			nextPageToken = ""
		}
		startNo := (i-1)*groupPerPage + 1
		groupSizeInPage := min(groupPerPage, totalGroups-startNo+1)

		if i == 1 {
			// for first page, pageToken is empty
			groupData[""] = googleGetGroups(nextPageToken, startNo, groupSizeInPage)
		} else {
			groupData[strconv.Itoa(i)] = googleGetGroups(nextPageToken, startNo, groupSizeInPage)
		}
	}

	return func(u *url.URL) (int, []byte) {
		if pgToken, ok := u.Query()["pageToken"]; ok {
			if len(pgToken) != 1 {
				return http.StatusBadRequest, []byte("invalid query page parameter")
			}
			data, err := json.Marshal(groupData[pgToken[0]])
			if err != nil {
				return http.StatusBadRequest, []byte(err.Error())
			}
			return http.StatusOK, data
		}

		return http.StatusBadRequest, []byte("invalid query page parameter")
	}
}

// groups search url parameter
// domain, pageToken, userKey
//
// domain, userKey must be non empty
func googleVerifyUrlParams(u *url.URL) error {
	urlParams := u.Query()
	queryParams := []string{"domain", "pageToken", "userKey"}

	for _, key := range queryParams {
		if val, ok := urlParams[key]; ok {
			if len(val) != 1 {
				return fmt.Errorf("invalid query %v parameter", key)
			}
			if key != "pageToken" && len(val[0]) == 0 {
				return fmt.Errorf("invalid query %v parameter value", key)
			}
		}
	}
	return nil
}

func googleClientSetup(serverUrl string) (*GoogleClient, error) {
	g := &GoogleClient{
		ctx: context.Background(),
		GoogleOptions: GoogleOptions{
			AdminEmail:             adminEmail,
			ServiceAccountJsonFile: "sa.json",
		},
	}
	p, err := oidc.NewProvider(g.ctx, serverUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for azure. Reason: %v", err)
	}

	g.verifier = p.Verifier(&oidc.Config{
		ClientID:        GoogleOauth2ClientID,
		SkipExpiryCheck: true,
	})

	g.service, err = gdir.New(http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create google service. Reason: %v", err)
	}
	g.service.BasePath = serverUrl
	g.service.Groups = gdir.NewGroupsService(g.service)

	return g, nil
}

func googleServerSetup(jwkResp []byte, groupResp googleGroupResp) (*httptest.Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return nil, err
	}
	addr := listener.Addr().String()

	m := pat.New()

	m.Get("/groups", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// have to work on error response
		err := googleVerifyUrlParams(r.URL)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		status, resp := groupResp(r.URL)
		w.WriteHeader(status)
		w.Write(resp)
	}))

	m.Get("/.well-known/openid-configuration", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := `{"issuer" : "http://%v", "jwks_uri" : "http://%v/jwk"}`
		w.Write([]byte(fmt.Sprintf(resp, addr, addr)))
	}))

	m.Get("/jwk", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(jwkResp)
	}))

	srv := &httptest.Server{
		Listener: listener,
		Config:   &http.Server{Handler: m},
	}
	srv.Start()

	return srv, nil
}

func googleVerifyGroups(groups []string, expectedSize int) error {
	if len(groups) != expectedSize {
		return fmt.Errorf("expected group size: %v, got %v", expectedSize, len(groups))
	}
	mapGroupName := map[string]bool{}
	for _, name := range groups {
		mapGroupName[name] = true
	}
	for i := 1; i <= expectedSize; i++ {
		group := googleGetGroupEmail(i)
		if _, ok := mapGroupName[group]; !ok {
			return fmt.Errorf("group %v is missing", group)
		}
	}
	return nil
}

func googleVerifyAuthenticatedReview(review auth.TokenReview, groupSize int) error {
	if !review.Status.Authenticated {
		return fmt.Errorf("expected authenticated ture, got false")
	}
	if review.Status.User.Username != userEmail {
		return fmt.Errorf("expected username %v, got %v", userEmail, review.Status.User.Username)
	}
	err := googleVerifyGroups(review.Status.User.Groups, groupSize)
	if err != nil {
		return err
	}
	return nil
}

func googleVerifyUnauthenticatedReview(review auth.TokenReview) error {
	if review.Status.Authenticated {
		return fmt.Errorf("expected authenticated false, got true")
	}
	if review.Status.Error == "" {
		return fmt.Errorf("expected error non empty")
	}
	return nil
}

func TestCheckGoogleAuthenticationSuccess(t *testing.T) {
	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	jwkSet := signKey.jwk()
	jwkResp, err := json.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	dataset := []struct {
		groupSize    int
		groupPerPage int
	}{
		{0, 5},
		{1, 5},
		{10, 5},
		{13, 5},
	}

	for _, test := range dataset {
		t.Run(fmt.Sprintf("authentication successful, group size: %v, group per page: %v", test.groupSize, test.groupPerPage), func(t *testing.T) {
			srv, err := googleServerSetup(jwkResp, googleGetGroupResp(test.groupSize, test.groupPerPage, numberOfPage(test.groupSize, test.groupPerPage)))
			if err != nil {
				t.Fatalf("Error when creating server, reason: %v", err)
			}
			defer srv.Close()

			client, err := googleClientSetup(srv.URL)
			if err != nil {
				t.Fatalf("Error when creatidng google client. reason : %v", err)
			}

			token, err := signKey.sign([]byte(fmt.Sprintf(googleToken, srv.URL, userEmail, GoogleOauth2ClientID, domain)))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}
			resp, status := client.checkGoogle(domain, token)
			if status != http.StatusOK {
				t.Errorf("Expected %v, got %v. reason: %v", http.StatusOK, status, resp.Status.Error)
			}
			err = googleVerifyAuthenticatedReview(resp, test.groupSize)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCheckGoogleAuthenticationFailed(t *testing.T) {

	var (
		emptyHDToken     = fmt.Sprintf(`{ "iss" : "%%v", "email" : "%v", "aud" : "%v"}`, userEmail, GoogleOauth2ClientID)
		badClientIDToken = fmt.Sprintf(`{ "iss" : "%%v", "email" : "%v", "aud" : "bad_client_id", "hd" : "%v"}`, userEmail, domain)
		badDomainIDToken = fmt.Sprintf(`{ "iss" : "%%v", "email" : "%v", "aud" : "%v", "hd" : "bad_domain"}`, userEmail, GoogleOauth2ClientID)
		goodIDToken      = fmt.Sprintf(`{ "iss" : "%%v", "email" : "%v", "aud" : "%v", "hd" : "%v"}`, userEmail, GoogleOauth2ClientID, domain)
	)

	signKey, err := newRSAKey()
	if err != nil {
		t.Fatalf("Error when creating signing key. reason : %v", err)
	}

	jwkSet := signKey.jwk()
	jwkResp, err := json.Marshal(jwkSet)
	if err != nil {
		t.Fatalf("Error when generating JSONWebKeySet. reason: %v", err)
	}

	groupErrResp := func(u *url.URL) (int, []byte) {
		return http.StatusInternalServerError, []byte(`{"message" : "error"}`)
	}

	dataset := []struct {
		testName  string
		token     string
		groupResp googleGroupResp
	}{
		{
			"authentication unsuccessful, reason invalid token",
			badToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason bad client ID",
			badClientIDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason empty hd(host domain)",
			emptyHDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason bad domain",
			badDomainIDToken,
			googleGetGroupResp(4, 5, 1),
		},
		{
			"authentication unsuccessful, reason error occurred in Page 1 (first page)",
			goodIDToken,
			groupErrResp,
		},
		{
			"authentication unsuccessful, reason error occurred in Page 2 (last page)",
			goodIDToken,
			groupErrResp,
		},
	}

	for _, test := range dataset {
		t.Run(test.testName, func(t *testing.T) {
			srv, err := googleServerSetup(jwkResp, test.groupResp)
			if err != nil {
				t.Fatalf("Error when creating server, reason: %v", err)
			}
			defer srv.Close()

			client, err := googleClientSetup(srv.URL)
			if err != nil {
				t.Fatalf("Error when creatidng google client. reason : %v", err)
			}

			token, err := signKey.sign([]byte(fmt.Sprintf(test.token, srv.URL)))
			if err != nil {
				t.Fatalf("Error when signing token. reason: %v", err)
			}
			resp, status := client.checkGoogle(domain, token)
			if status != http.StatusUnauthorized {
				t.Errorf("Expected %v, got %v", http.StatusUnauthorized, status)
			}
			err = googleVerifyUnauthenticatedReview(resp)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
