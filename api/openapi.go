// Package api contains API definitions
package api

import _ "embed"

// OpenAPISpec is the OpenAPI specification of HTTP API
//
//go:embed http/openapi.yaml
var OpenAPISpec []byte
