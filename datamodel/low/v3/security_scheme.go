// Copyright 2022 Princess B33f Heavy Industries / Dave Shanley
// SPDX-License-Identifier: MIT

package v3

import (
	"crypto/sha256"
	"fmt"
	"github.com/pb33f/libopenapi/datamodel/low"
	"github.com/pb33f/libopenapi/index"
	"gopkg.in/yaml.v3"
	"strings"
)

// SecurityScheme represents a low-level OpenAPI 3+ SecurityScheme object.
//
// Defines a security scheme that can be used by the operations.
//
// Supported schemes are HTTP authentication, an API key (either as a header, a cookie parameter or as a query parameter),
// mutual TLS (use of a client certificate), OAuth2’s common flows (implicit, password, client credentials and
// authorization code) as defined in RFC6749 (https://www.rfc-editor.org/rfc/rfc6749), and OpenID Connect Discovery.
// Please note that as of 2020, the implicit  flow is about to be deprecated by OAuth 2.0 Security Best Current Practice.
// Recommended for most use case is Authorization Code Grant flow with PKCE.
//  - https://spec.openapis.org/oas/v3.1.0#security-scheme-object
type SecurityScheme struct {
	Type             low.NodeReference[string]
	Description      low.NodeReference[string]
	Name             low.NodeReference[string]
	In               low.NodeReference[string]
	Scheme           low.NodeReference[string]
	BearerFormat     low.NodeReference[string]
	Flows            low.NodeReference[*OAuthFlows]
	OpenIdConnectUrl low.NodeReference[string]
	Extensions       map[low.KeyReference[string]]low.ValueReference[any]
}

// SecurityRequirement is a low-level representation of an OpenAPI 3+ SecurityRequirement object.
//
// It lists the required security schemes to execute this operation. The name used for each property MUST correspond
// to a security scheme declared in the Security Schemes under the Components Object.
//
// Security Requirement Objects that contain multiple schemes require that all schemes MUST be satisfied for a
// request to be authorized. This enables support for scenarios where multiple query parameters or HTTP headers are
// required to convey security information.
//
// When a list of Security Requirement Objects is defined on the OpenAPI Object or Operation Object, only one of the
// Security Requirement Objects in the list needs to be satisfied to authorize the request.
//  - https://spec.openapis.org/oas/v3.1.0#security-requirement-object
//type SecurityRequirement struct {
//
//	// FYI, I hate this data structure. Even without the low level wrapping, it sucks.
//	ValueRequirements []low.ValueReference[map[low.KeyReference[string]]low.ValueReference[[]low.ValueReference[string]]]
//}

// FindExtension attempts to locate an extension using the supplied key.
func (ss *SecurityScheme) FindExtension(ext string) *low.ValueReference[any] {
	return low.FindItemInMap[any](ext, ss.Extensions)
}

// Build will extract OAuthFlows and extensions from the node.
func (ss *SecurityScheme) Build(root *yaml.Node, idx *index.SpecIndex) error {
	ss.Extensions = low.ExtractExtensions(root)

	oa, oaErr := low.ExtractObject[*OAuthFlows](OAuthFlowsLabel, root, idx)
	if oaErr != nil {
		return oaErr
	}
	if oa.Value != nil {
		ss.Flows = oa
	}
	return nil
}

// Hash will return a consistent SHA256 Hash of the SecurityScheme object
func (ss *SecurityScheme) Hash() [32]byte {
	var f []string
	if !ss.Type.IsEmpty() {
		f = append(f, ss.Type.Value)
	}
	if !ss.Description.IsEmpty() {
		f = append(f, ss.Description.Value)
	}
	if !ss.Name.IsEmpty() {
		f = append(f, ss.Name.Value)
	}
	if !ss.In.IsEmpty() {
		f = append(f, ss.In.Value)
	}
	if !ss.Scheme.IsEmpty() {
		f = append(f, ss.Scheme.Value)
	}
	if !ss.BearerFormat.IsEmpty() {
		f = append(f, ss.BearerFormat.Value)
	}
	if !ss.Flows.IsEmpty() {
		f = append(f, low.GenerateHashString(ss.Flows.Value))
	}
	if !ss.OpenIdConnectUrl.IsEmpty() {
		f = append(f, ss.OpenIdConnectUrl.Value)
	}
	for k := range ss.Extensions {
		f = append(f, fmt.Sprintf("%s-%v", k.Value, ss.Extensions[k].Value))
	}
	return sha256.Sum256([]byte(strings.Join(f, "|")))
}
