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

package graph

import (
	"net/http"
)

type passthrough struct {
	name   string
	client *http.Client
}

func NewPassthroughTokenProvider() TokenProvider {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	return &passthrough{
		name:   "PassthroughTokenProvider",
		client: &http.Client{Transport: tr},
	}
}

func (u *passthrough) Name() string { return u.name }

func (u *passthrough) Acquire(token string) (AuthResponse, error) {
	return AuthResponse{Token: token}, nil
}
