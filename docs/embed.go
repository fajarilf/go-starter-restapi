package docs

import "embed"

// FS holds the embedded OpenAPI specification files served at runtime
// (openapi.yaml plus any files it $refs, e.g. room_docs.yaml).
//
//go:embed openapi.yaml room_docs.yaml
var FS embed.FS
