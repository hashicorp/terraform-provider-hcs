// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package helper

import (
	"fmt"
	"reflect"

	"github.com/dgrijalva/jwt-go"
)

// federationTokenClaims represents the JWT claims for an HCS Consul federation token.
// We don't need to deserialize the Primary link since we are only using it to diff
// against other federation tokens for the same cluster.
type federationTokenClaims struct {
	jwt.StandardClaims
	Primary interface{} `json:"Primary"`
}

// Valid is required to implement the JWT Claims interface.
// We are not validating the claims here since we only care about the 'Primary'
// field on the JWT payload and all validation is handled in the HCS API.
func (_ federationTokenClaims) Valid() error {
	return nil
}

// FederationTokensHaveSamePrimary determines if two federation tokens (base64 encoded JWTs)
// have the same 'Primary' claim, which denotes that the tokens belong to the same primary HCS
// cluster.
func FederationTokensHaveSamePrimary(encodedToken1, encodedToken2 string) bool {
	claims1, err := extractEncodedFederationTokenClaims(encodedToken1)
	if err != nil || claims1 == nil {
		return false
	}

	claims2, err := extractEncodedFederationTokenClaims(encodedToken2)
	if err != nil || claims2 == nil {
		return false
	}

	return reflect.DeepEqual(&claims1.Primary, &claims2.Primary)
}

// extractEncodedFederationTokenClaims extracts a pointer of federationTokenClaims
// from an encoded JWT string.
func extractEncodedFederationTokenClaims(token string) (*federationTokenClaims, error) {
	parser := new(jwt.Parser)

	parsedToken, _, err := parser.ParseUnverified(token, &federationTokenClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*federationTokenClaims)
	if !ok {
		return nil, fmt.Errorf("unable to extract claims from federation token (type assertion failure)")
	}

	return claims, nil
}
