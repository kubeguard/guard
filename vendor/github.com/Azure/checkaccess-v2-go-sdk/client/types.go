package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

// AuthorizationRequest represents the payload of the request sent to a PDP server
type AuthorizationRequest struct {
	Subject            SubjectInfo     `json:"Subject"`
	Actions            []ActionInfo    `json:"Actions"`
	Resource           ResourceInfo    `json:"Resource"`
	Environment        EnvironmentInfo `json:"Environment,omitempty"`
	CheckClassicAdmins bool            `json:"CheckClassicAdmins,omitempty"`
}

type SubjectInfo struct {
	Attributes SubjectAttributes `json:"Attributes"`
}

// SubjectAttributes contains the possible attributes to describe the subject
// of query (i.e. if IT has the access). The ObjectId field is the UUID value of
// the subject and is required.
type SubjectAttributes struct {
	ObjectId         string   `json:"ObjectId"`
	Groups           []string `json:"Groups,omitempty"`
	ApplicationId    string   `json:"ApplicationId,omitempty"`
	ApplicationACR   string   `json:"ApplicationACR,omitempty"`
	RoleTemplate     []string `json:"RoleTemplate,omitempty"`
	TenantId         string   `json:"tid,omitempty"`
	Scope            string   `json:"Scope,omitempty"`
	ResourceId       string   `json:"ResourceId,omitempty"`
	Puid             string   `json:"puid,omitempty"`
	AltSecId         string   `json:"altsecid,omitempty"`
	IdentityProvider string   `json:"idp,omitempty"`
	Issuer           string   `json:"iss,omitempty"`
	ClaimName        string   `json:"_claim_names,omitempty"`
}

// ActionInfo contains an action the query checks whether the subject
// has access to perform. Example: "Microsoft.Network/virtualNetworks/read"
type ActionInfo struct {
	Id           string `json:"Id"`
	IsDataAction bool   `json:"IsDataAction,omitempty"`
	Attributes   `json:"Attributes"`
}

// ResourceInfo is the resource path of the target object the query
// checks whether the subject has access to perform against it.
type ResourceInfo struct {
	Id         string `json:"Id"`
	Attributes `json:"Attributes"`
}

type EnvironmentInfo struct {
	Attributes `json:"Attributes"`
}

// AuthorizationDecisionResponse contains a paginated list of all decision results
// In case the list is more than 50, follow NextLink to retrieve the next page.
type AuthorizationDecisionResponse struct {
	Value    []AuthorizationDecision `json:"value"`
	NextLink string                  `json:"nextLink"`
}

// AuthorizationDecision tells whether the subject can perform the action
// on the target resource.
type AuthorizationDecision struct {
	ActionId       string `json:"actionId,omitempty"`
	AccessDecision `json:"accessDecision,omitempty"`
	IsDataAction   bool `json:"isDataAction,omitempty"`
	RoleAssignment `json:"roleAssignment,omitempty"`
	DenyAssignment RoleDefinition `json:"denyAssignment,omitempty"`
	TimeToLiveInMs int            `json:"timeToLiveInMs,omitempty"`
}

type RoleAssignment struct {
	Id                                 string `json:"id,omitempty"`
	RoleDefinitionId                   string `json:"roleDefinitionId,omitempty"`
	PrincipalId                        string `json:"principalId,omitempty"`
	PrincipalType                      string `json:"principaltype,omitempty"`
	Scope                              string `json:"scope,omitempty"`
	Condition                          string `json:"condition,omitempty"`
	ConditionVersion                   string `json:"conditionVersion,omitempty"`
	CanDelegate                        bool   `json:"canDelegate,omitempty"`
	DelegatedManagedIdentityResourceId string `json:"deletegatedManagedIdentityResourceId,omitempty"`
	Description                        string `json:"description,omitempty"`
}

type RoleDefinition struct {
	Id string `json:"id,omitempty"`
}

type Attributes map[string]interface{}

// RemotePDPErrorPayload represents the body content when the server returns
// a non-successful error
type CheckAccessErrorResponse struct {
	StatusCode int    `json:"statusCode,omitempty"`
	Message    string `json:"message,omitempty"`
}
