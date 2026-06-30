package runtimebin

import "testing"

func TestDetectKnownBinary(t *testing.T) {
	d := Detector{}

	path, found := d.Detect("go")
	if !found {
		t.Fatalf("Detect(go) found = false, want true")
	}
	if path == "" {
		t.Errorf("Detect(go) path = empty, want non-empty")
	}

	if _, found := d.Detect("definitely-not-a-real-binary-xyz123"); found {
		t.Errorf("Detect(bogus) found = true, want false")
	}
}
