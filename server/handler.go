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

package server

import (
	"context"
	"net/http"
	"strings"

	"go.kubeguard.dev/guard/auth"
	"go.kubeguard.dev/guard/auth/providers/azure"
	"go.kubeguard.dev/guard/auth/providers/github"
	"go.kubeguard.dev/guard/auth/providers/gitlab"
	"go.kubeguard.dev/guard/auth/providers/google"
	"go.kubeguard.dev/guard/auth/providers/ldap"
	"go.kubeguard.dev/guard/auth/providers/token"
	errutils "go.kubeguard.dev/guard/util/error"

	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	"k8s.io/klog/v2"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		write(w, nil, errutils.WithCode(errors.New("Missing client certificate"), http.StatusBadRequest))
		return
	}
	crt := req.TLS.PeerCertificates[0]
	if len(crt.Subject.Organization) == 0 {
		write(w, nil, errutils.WithCode(errors.New("Client certificate is missing organization"), http.StatusBadRequest))
		return
	}
	org := crt.Subject.Organization[0]
	klog.V(7).Infof("Received token review request for %s/%s", org, crt.Subject.CommonName)

	data := authv1.TokenReview{}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		write(w, nil, errutils.WithCode(errors.Wrap(err, "Failed to parse request"), http.StatusBadRequest))
		return
	}

	if !s.AuthRecommendedOptions.AuthProvider.Has(org) {
		write(w, nil, errutils.WithCode(errors.Errorf("guard does not provide service for %v", org), http.StatusBadRequest))
		return
	}

	if s.AuthRecommendedOptions.AuthProvider.Has(token.OrgType) && s.TokenAuthenticator != nil {
		resp, err := s.TokenAuthenticator.Check(data.Spec.Token)
		if err == nil {
			write(w, resp, err)
			return
		}
	}

	ctx := req.Context()
	client, err := s.getAuthProviderClient(ctx, org, crt.Subject.CommonName)
	if err != nil {
		write(w, nil, err)
		return
	}

	resp, err := client.Check(ctx, data.Spec.Token)
	write(w, resp, err)
}

func (s *Server) getAuthProviderClient(ctx context.Context, org, commonName string) (auth.Interface, error) {
	switch strings.ToLower(org) {
	case github.OrgType:
		return github.New(s.AuthRecommendedOptions.Github, commonName), nil
	case google.OrgType:
		return google.New(ctx, s.AuthRecommendedOptions.Google, commonName)
	case gitlab.OrgType:
		return gitlab.New(s.AuthRecommendedOptions.Gitlab), nil
	case azure.OrgType:
		return azure.New(ctx, s.AuthRecommendedOptions.Azure)
	case ldap.OrgType:
		return ldap.New(s.AuthRecommendedOptions.LDAP), nil
	}

	return nil, errors.Errorf("Client is using unknown organization %s", org)
}
