//go:build linux && amd64

package internaltools

import _ "embed"

// The build script replaces this placeholder with the statically linked standalone vi binary.
//go:embed assets/vi
var viPayload []byte
