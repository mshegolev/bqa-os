package knowledge

import (
	"context"
	"testing"
)

type mapReader struct{ files map[string]string }

func (m mapReader) ReadKnowledgeArtifact(_ context.Context, name string) (string, error) {
	c, ok := m.files[name]
	if !ok {
		return "", errMissing
	}
	return c, nil
}

var errMissing = &missingErr{}

type missingErr struct{}

func (*missingErr) Error() string { return "missing" }

// buildValidFiles renders a full, valid v1 artifact set for the given profile.
func buildValidFiles() map[string]string {
	files := map[string]string{}
	for _, spec := range ExpectedArtifacts() {
		if spec.RootKey == "project_profile" {
			files[spec.Filename] = renderProfile(ProjectProfile{Sessions: 3, ETLSignals: 1})
		} else {
			files[spec.Filename] = renderFindings(spec.RootKey, nil)
		}
	}
	return files
}

func TestValidateAcceptsV1(t *testing.T) {
	rep := Validate(context.Background(), mapReader{files: buildValidFiles()})
	if !rep.OK() {
		t.Fatalf("expected valid v1 set, got issues: %+v", rep.Issues)
	}
}

func TestValidateRejectsMissingSchemaVersion(t *testing.T) {
	files := buildValidFiles()
	files["etl_patterns.yaml"] = "kind: etl_patterns\npatterns: []\n" // no schema_version
	rep := Validate(context.Background(), mapReader{files: files})
	if rep.OK() {
		t.Fatalf("expected failure for missing schema_version")
	}
}

func TestValidateRejectsWrongKind(t *testing.T) {
	files := buildValidFiles()
	files["etl_patterns.yaml"] = "schema_version: 1\nkind: wrong\npatterns: []\n"
	rep := Validate(context.Background(), mapReader{files: files})
	if rep.OK() {
		t.Fatalf("expected failure for wrong kind")
	}
}
