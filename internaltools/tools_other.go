//go:build !linux || !amd64

package internaltools

var (
	busyboxPayload []byte
	lessPayload    []byte
)
