//go:build linux && amd64

package internaltools

import _ "embed"

// The build script replaces these placeholders with pinned upstream binaries.
//go:embed assets/busybox
var busyboxPayload []byte

//go:embed assets/less
var lessPayload []byte
