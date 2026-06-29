package fs

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestDemoArchiveWriterCreatesDeterministicZip(t *testing.T) {
	tmp := t.TempDir()
	files := []ports.DemoArchiveFile{
		{Path: "z-last.md", Content: "Synthetic last"},
		{Path: "a-first.md", Content: "Synthetic first"},
	}

	outputA := filepath.Join(tmp, "demo-a.zip")
	outputB := filepath.Join(tmp, "nested", "demo-b.zip")
	writer := DemoArchiveWriter{}

	if err := writer.WriteDemoArchive(context.Background(), outputA, files); err != nil {
		t.Fatalf("WriteDemoArchive A returned error: %v", err)
	}
	if err := writer.WriteDemoArchive(context.Background(), outputB, files); err != nil {
		t.Fatalf("WriteDemoArchive B returned error: %v", err)
	}

	dataA, err := os.ReadFile(outputA)
	if err != nil {
		t.Fatalf("ReadFile A returned error: %v", err)
	}
	dataB, err := os.ReadFile(outputB)
	if err != nil {
		t.Fatalf("ReadFile B returned error: %v", err)
	}
	if !bytes.Equal(dataA, dataB) {
		t.Fatalf("expected deterministic zip bytes for same inputs")
	}

	reader, err := zip.OpenReader(outputA)
	if err != nil {
		t.Fatalf("OpenReader returned error: %v", err)
	}
	defer reader.Close()

	var names []string
	for _, file := range reader.File {
		names = append(names, file.Name)
		if !file.Modified.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
			t.Fatalf("file %s has non-deterministic modified time %s", file.Name, file.Modified)
		}
	}

	expected := []string{"a-first.md", "z-last.md"}
	if !reflect.DeepEqual(names, expected) {
		t.Fatalf("zip entries = %#v, expected %#v", names, expected)
	}
}
