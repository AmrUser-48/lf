//go:build linux && amd64

package internaltools

import _ "embed"

// The build script replaces this placeholder with the pinned upstream BusyBox binary.
//go:embed assets/busybox
var busyboxPayload []byte
