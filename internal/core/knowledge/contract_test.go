package knowledge

import "testing"

func TestSchemaVersionIsOne(t *testing.T) {
	if SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", SchemaVersion)
	}
}

func TestExpectedArtifactsRootKeysServeAsKind(t *testing.T) {
	for _, spec := range ExpectedArtifacts() {
		if spec.RootKey == "" {
			t.Fatalf("artifact %q has empty RootKey (used as kind)", spec.Filename)
		}
	}
}
