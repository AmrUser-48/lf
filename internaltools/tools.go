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

// Vi returns the bundled standalone C editor when its payload is embedded.
// Builds without the payload fail explicitly instead of launching a system editor.
func Vi(filename string) (*exec.Cmd, func(), error) {
	if bytes.HasPrefix(viPayload, []byte{0x7f, 'E', 'L', 'F'}) {
		path, cleanup, err := materialize(viPayload)
		if err != nil {
			return nil, noCleanup, err
		}
		return exec.Command(path, filename), cleanup, nil
	}

	return nil, noCleanup, fmt.Errorf("embedded standalone vi is unavailable in this build; run scripts/build-vi.sh on Linux amd64 and rebuild lf")
}

func materialize(payload []byte) (string, func(), error) {
	digest := sha256.Sum256(payload)
	dir, err := os.MkdirTemp("", fmt.Sprintf("lf-vi-%x-", digest[:6]))
	if err != nil {
		return "", noCleanup, fmt.Errorf("create temporary vi directory: %w", err)
	}
	path := filepath.Join(dir, "vi")
	if err := os.WriteFile(path, payload, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("write bundled vi: %w", err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("make bundled vi executable: %w", err)
	}
	return path, func() { _ = os.RemoveAll(dir) }, nil
}
