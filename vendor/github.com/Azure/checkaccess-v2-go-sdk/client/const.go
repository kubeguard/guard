package client

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

const (
	modulename = "aro-pdpclient"
	// version is the semantic version of this module
	version = "0.0.1"
)

// AccessDecision can be: Allowed, NotAllowed, Denied.
type AccessDecision string

// AccessDecision possible returned values
const (
	Allowed    AccessDecision = "Allowed"
	NotAllowed AccessDecision = "NotAllowed"
	Denied     AccessDecision = "Denied"

	// GroupExpansion is the value to be used with ClaimName in SubjectAttributes
	// This value gives CheckAccess a hint that it needs to retrieve all the groups the principal belongs to
	// and then give the response based on all group entitlements.
	//
	// https://eng.ms/docs/microsoft-security/identity/auth/access-control-managed-identityacmi/azure-authz-data-plane/authz-dataplane-partner-wiki/remotepdp/checkaccess/samples/requestresponse
	GroupExpansion = `{"groups":"src1"}`
)
