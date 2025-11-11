package token

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

import (
	"github.com/Azure/checkaccess-v2-go-sdk/client/internal"
	"github.com/golang-jwt/jwt/v4"
)

// ExtractClaims extracts the "oid", "_claim_names", and "groups" claims from a given access jwtToken and return them as a custom struct
func ExtractClaims(jwtToken string) (*internal.Custom, error) {
	p := jwt.NewParser(jwt.WithoutClaimsValidation())
	c := &internal.Custom{}
	_, _, err := p.ParseUnverified(jwtToken, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
