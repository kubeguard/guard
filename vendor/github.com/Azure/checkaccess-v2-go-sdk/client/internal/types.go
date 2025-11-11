package internal

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.
import (
	"github.com/golang-jwt/jwt/v4"
)

type Custom struct {
	ObjectId   string                 `json:"oid"`
	ClaimNames map[string]interface{} `json:"_claim_names"`
	Groups     []string               `json:"groups"`
	jwt.RegisteredClaims
}
