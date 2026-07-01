package fs

import "testing"

func TestCleanArchivePathRejectsTraversal(t *testing.T) {
	if _, err := cleanArchivePath("../evil"); err == nil {
		t.Fatal("expected error for traversal path")
	}
	if _, err := cleanArchivePath("/abs"); err == nil {
		t.Fatal("expected error for absolute path")
	}
	got, err := cleanArchivePath("knowledge\\etl.yaml")
	if err != nil || got != "knowledge/etl.yaml" {
		t.Fatalf("cleanArchivePath normalize = %q, %v", got, err)
	}
}
