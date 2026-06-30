// Package registry reads the BQA unified registry (team/brain/registry.json)
// and the raw source files of the artifacts it references.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mshegolev/bqa-os/internal/pathsafe"
	"github.com/mshegolev/bqa-os/internal/ports"
)

const expectedKind = "BQATeamUnifiedRegistry"

// Reader loads the registry JSON at RegistryPath and resolves artifact source
// files relative to Root.
type Reader struct {
	RegistryPath string
	Root         string
}

type jsonRegistry struct {
	Version  int    `json:"version"`
	Kind     string `json:"kind"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Artifacts []jsonArtifact `json:"artifacts"`
}

type jsonArtifact struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Summary     string   `json:"summary"`
	Tags        []string `json:"tags"`
}

func (r Reader) LoadRuntimeRegistry(ctx context.Context) (ports.RuntimeRegistry, error) {
	select {
	case <-ctx.Done():
		return ports.RuntimeRegistry{}, ctx.Err()
	default:
	}

	data, err := os.ReadFile(r.RegistryPath)
	if err != nil {
		return ports.RuntimeRegistry{}, err
	}

	var parsed jsonRegistry
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ports.RuntimeRegistry{}, fmt.Errorf("parse registry %s: %w", r.RegistryPath, err)
	}
	if parsed.Kind != expectedKind {
		return ports.RuntimeRegistry{}, fmt.Errorf("unsupported registry kind %q (want %q)", parsed.Kind, expectedKind)
	}

	registry := ports.RuntimeRegistry{
		Version: parsed.Version,
		Kind:    parsed.Kind,
		Name:    parsed.Metadata.Name,
	}
	seen := map[string]bool{}
	for _, artifact := range parsed.Artifacts {
		if artifact.ID == "" {
			return ports.RuntimeRegistry{}, fmt.Errorf("registry contains an artifact with an empty id")
		}
		if seen[artifact.ID] {
			return ports.RuntimeRegistry{}, fmt.Errorf("duplicate artifact id: %s", artifact.ID)
		}
		seen[artifact.ID] = true
		registry.Artifacts = append(registry.Artifacts, ports.RuntimeArtifact{
			ID:          artifact.ID,
			Type:        artifact.Type,
			Title:       artifact.Title,
			Source:      artifact.Source,
			Destination: artifact.Destination,
			Summary:     artifact.Summary,
			Tags:        artifact.Tags,
		})
	}
	return registry, nil
}

func (r Reader) ReadArtifactSource(ctx context.Context, source string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	cleaned, ok := pathsafe.RelClean(source)
	if !ok {
		return "", fmt.Errorf("unsafe artifact source: %s", source)
	}
	data, err := os.ReadFile(filepath.Join(r.Root, cleaned))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
