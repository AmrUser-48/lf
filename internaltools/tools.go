package internaltools

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var noCleanup = func() {}

// Vi returns the bundled upstream BusyBox vi applet when a Linux amd64 release
// payload is available; source builds fall back to a system BusyBox executable.
func Vi(filename string) (*exec.Cmd, func(), error) {
	if bytes.HasPrefix(busyboxPayload, []byte{0x7f, 'E', 'L', 'F'}) {
		path, cleanup, err := materialize(busyboxPayload)
		if err != nil {
			return nil, noCleanup, err
		}
		return exec.Command(path, "vi", filename), cleanup, nil
	}

	path, err := exec.LookPath("busybox")
	if err != nil {
		return nil, noCleanup, fmt.Errorf("BusyBox vi payload is not built; run scripts/build-internal-tools.sh or install busybox in PATH: %w", err)
	}
	return exec.Command(path, "vi", filename), noCleanup, nil
}

func materialize(payload []byte) (string, func(), error) {
	digest := sha256.Sum256(payload)
	dir, err := os.MkdirTemp("", fmt.Sprintf("lf-busybox-vi-%x-", digest[:6]))
	if err != nil {
		return "", noCleanup, fmt.Errorf("create temporary BusyBox directory: %w", err)
	}
	path := filepath.Join(dir, "busybox")
	if err := os.WriteFile(path, payload, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("write bundled BusyBox: %w", err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("make bundled BusyBox executable: %w", err)
	}
	return path, func() { _ = os.RemoveAll(dir) }, nil
}
