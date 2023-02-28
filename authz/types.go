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
package authz

import (
	"context"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	authzv1 "k8s.io/api/authorization/v1"
)

type orgs []string

var SupportedOrgs orgs

func (o orgs) Has(name string) bool {
	name = strings.TrimSpace(name)
	for _, org := range o {
		if strings.EqualFold(name, org) {
			return true
		}
	}
	return false
}

func (o orgs) String() string {
	names := make([]string, len(o))
	for i, org := range o {
		names[i] = cases.Title(language.English).String(org)
	}
	sort.Strings(names)
	return strings.Join(names, "/")
}

type Interface interface {
	Check(ctx context.Context, request *authzv1.SubjectAccessReviewSpec, store Store) (*authzv1.SubjectAccessReviewStatus, error)
}

type Store interface {
	Set(key string, value interface{}) error
	Get(key string, value interface{}) (bool, error)
	Delete(key string) error
	Close() error
}
