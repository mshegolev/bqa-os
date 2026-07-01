package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBrainExportThenImportRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	bqa := filepath.Join(tmp, ".bqa")
	for rel, content := range map[string]string{
		"knowledge/etl.yaml":  "schema_version: 1\n",
		"registry/index.yaml": "registry:\n  version: 1\n",
	} {
		p := filepath.Join(bqa, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	bundle := filepath.Join(tmp, "bundle.zip")

	// export
	exp := brainCmd()
	var out bytes.Buffer
	exp.SetOut(&out)
	exp.SetErr(&out)
	exp.SetArgs([]string{"export", "--source", bqa, "--target", "zip", "--out", bundle})
	if err := exp.Execute(); err != nil {
		t.Fatalf("export: %v\n%s", err, out.String())
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Fatalf("bundle not written: %v", err)
	}

	// import
	target := filepath.Join(tmp, "client")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	imp := brainCmd()
	var out2 bytes.Buffer
	imp.SetOut(&out2)
	imp.SetErr(&out2)
	imp.SetArgs([]string{"import", "--from", bundle, "--target", target})
	if err := imp.Execute(); err != nil {
		t.Fatalf("import: %v\n%s", err, out2.String())
	}
	if _, err := os.Stat(filepath.Join(target, ".bqa", "knowledge", "etl.yaml")); err != nil {
		t.Fatalf("expected installed file: %v", err)
	}
}
