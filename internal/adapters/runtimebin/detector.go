// Package runtimebin detects AI coding runtime binaries on the host PATH.
package runtimebin

import "os/exec"

// Detector resolves runtime binaries via the host PATH.
type Detector struct{}

// Detect reports the absolute path of binary and whether it was found on PATH.
func (Detector) Detect(binary string) (string, bool) {
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", false
	}
	return path, true
}
