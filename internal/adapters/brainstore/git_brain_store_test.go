package brainstore

import (
	"context"
	"testing"

	"github.com/mshegolev/bqa-os/internal/ports"
)

func TestPushRequiresConnectedBrain(t *testing.T) {
	t.Setenv("BQA_HOME", t.TempDir()) // no config.yaml => not connected
	err := GitBrainStore{}.Push(context.Background(), []ports.ArchiveFile{{Path: "knowledge/x.yaml", Data: []byte("k")}}, false)
	if err == nil {
		t.Fatal("expected error when brain is not connected")
	}
}
