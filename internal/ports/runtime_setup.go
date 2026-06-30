package ports

// RuntimeDetector reports whether a runtime binary is available on the host
// and where it lives.
type RuntimeDetector interface {
	Detect(binary string) (path string, found bool)
}
