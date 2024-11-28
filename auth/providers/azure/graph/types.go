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

// AuthResponse represents a response from the MS Graph auth API
type AuthResponse struct {
	TokenType string `json:"token_type"`
	Expires   int    `json:"expires_in"`
	Token     string `json:"access_token"`
	// This is the actual time the token expires in Unix time
	ExpiresOn int `json:"expires_on"`
}

// NOTE: These below are partial implementations of the API objects containing
// only the necessary fields to perform the functions of this package

// ObjectList represents a list of directory object IDs returned from the MS Graph API
type ObjectList struct {
	Value []string `json:"value"`
}

// ObjectQuery represents a query object to the directoryObjects endpoint
type ObjectQuery struct {
	IDs   []string `json:"ids"`
	Types []string `json:"types"`
}

// GroupList represents a list of groups returned from the MS Graph API
type GroupList struct {
	Value []Group `json:"value"`
}

// Group represents the Group object from the MSGraphAPI
type Group struct {
	Name string `json:"displayName"`
	ID   string `json:"id"`
}
