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

func Command(tool, filename string) (*exec.Cmd, func(), error) {
	var payload []byte
	switch tool {
	case "busybox-vi":
		payload = busyboxPayload
	case "less":
		payload = lessPayload
	default:
		return nil, noCleanup, fmt.Errorf("unknown bundled tool %q", tool)
	}

	if bytes.HasPrefix(payload, []byte{0x7f, 'E', 'L', 'F'}) {
		path, cleanup, err := materialize(tool, payload)
		if err != nil {
			return nil, noCleanup, err
		}
		if tool == "busybox-vi" {
			return exec.Command(path, "vi", filename), cleanup, nil
		}
		return exec.Command(path, filename), cleanup, nil
	}

	// A source checkout without generated payloads can use upstream tools
	// installed on the host. Release builds embed the pinned upstream tools.
	var executable string
	var args []string
	switch tool {
	case "busybox-vi":
		executable = "busybox"
		args = []string{"vi", filename}
	case "less":
		executable = "less"
		args = []string{filename}
	}
	path, err := exec.LookPath(executable)
	if err != nil {
		return nil, noCleanup, fmt.Errorf("%s payload is not built; run scripts/build-internal-tools.sh (or install %s in PATH): %w", tool, executable, err)
	}
	return exec.Command(path, args...), noCleanup, nil
}

func materialize(name string, payload []byte) (string, func(), error) {
	digest := sha256.Sum256(payload)
	dir, err := os.MkdirTemp("", fmt.Sprintf("lf-%s-%x-", name, digest[:6]))
	if err != nil {
		return "", noCleanup, fmt.Errorf("create temporary tool directory: %w", err)
	}
	filename := name
	if name == "busybox-vi" {
		filename = "busybox"
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, payload, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("write bundled %s: %w", name, err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		_ = os.RemoveAll(dir)
		return "", noCleanup, fmt.Errorf("make bundled %s executable: %w", name, err)
	}
	return path, func() { _ = os.RemoveAll(dir) }, nil
}
