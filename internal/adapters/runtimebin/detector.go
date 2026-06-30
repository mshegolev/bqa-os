// Package runtimebin detects AI coding runtime binaries on the host PATH.
package runtimebin

import (
	"os/exec"

	"github.com/mshegolev/bqa-os/internal/ports"
)

// Detector resolves runtime binaries via the host PATH.
type Detector struct{}

var _ ports.RuntimeDetector = Detector{}

// Detect reports the absolute path of binary and whether it was found on PATH.
func (Detector) Detect(binary string) (string, bool) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", false
	}
	return path, true
}
