package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mshegolev/bqa-os/internal/ports"
)

type BQAWorkspaceStore struct {
	RootDir string
}

func (s BQAWorkspaceStore) LoadBQARegistry(ctx context.Context) (ports.BQARegistry, error) {
	if err := ctx.Err(); err != nil {
		return ports.BQARegistry{}, err
	}

	index, err := s.readWorkspaceFile(ctx, ".bqa/registry/index.yaml")
	if err != nil {
		return ports.BQARegistry{}, fmt.Errorf("read registry index: %w", err)
	}
	if !strings.Contains(index, "registry:") {
		return ports.BQARegistry{}, errors.New("registry index is missing registry root")
	}

	agents, err := s.loadRegistryItems(ctx, ".bqa/registry/agents.yaml", "agents")
	if err != nil {
		return ports.BQARegistry{}, err
	}
	skills, err := s.loadRegistryItems(ctx, ".bqa/registry/skills.yaml", "skills")
	if err != nil {
		return ports.BQARegistry{}, err
	}
	workflows, err := s.loadRegistryItems(ctx, ".bqa/registry/workflows.yaml", "workflows")
	if err != nil {
		return ports.BQARegistry{}, err
	}

	return ports.BQARegistry{
		Agents:    agents,
		Skills:    skills,
		Workflows: workflows,
	}, nil
}

func (s BQAWorkspaceStore) InspectWorkspacePath(ctx context.Context, path string) (ports.WorkspaceEntry, error) {
	if err := ctx.Err(); err != nil {
		return ports.WorkspaceEntry{}, err
	}

	cleanPath := cleanWorkspacePath(path)
	info, err := os.Stat(s.workspacePath(cleanPath))
	if os.IsNotExist(err) {
		return ports.WorkspaceEntry{Path: cleanPath}, nil
	}
	if err != nil {
		return ports.WorkspaceEntry{Path: cleanPath}, err
	}
	return ports.WorkspaceEntry{
		Path:   cleanPath,
		Exists: true,
		IsDir:  info.IsDir(),
		Size:   info.Size(),
	}, nil
}

func (s BQAWorkspaceStore) loadRegistryItems(ctx context.Context, path string, section string) ([]ports.BQARegistryItem, error) {
	content, err := s.readWorkspaceFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("read %s registry: %w", section, err)
	}
	items, err := parseRegistryItems(content, section)
	if err != nil {
		return nil, fmt.Errorf("parse %s registry: %w", section, err)
	}
	return items, nil
}

func (s BQAWorkspaceStore) readWorkspaceFile(ctx context.Context, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(s.workspacePath(path))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s BQAWorkspaceStore) workspacePath(path string) string {
	root := s.RootDir
	if root == "" {
		root = "."
	}
	return filepath.Join(root, filepath.FromSlash(cleanWorkspacePath(path)))
}

func cleanWorkspacePath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}

func parseRegistryItems(content string, section string) ([]ports.BQARegistryItem, error) {
	var items []ports.BQARegistryItem
	var current ports.BQARegistryItem
	inSection := false
	seenSection := false

	flush := func() error {
		if current == (ports.BQARegistryItem{}) {
			return nil
		}
		if current.ID == "" {
			return errors.New("item is missing id")
		}
		if current.Path == "" {
			return fmt.Errorf("item %q is missing path", current.ID)
		}
		items = append(items, current)
		current = ports.BQARegistryItem{}
		return nil
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if trimmed == section+":" {
			inSection = true
			seenSection = true
			continue
		}
		if !inSection {
			continue
		}
		if strings.HasPrefix(trimmed, "- id:") {
			if err := flush(); err != nil {
				return nil, err
			}
			current.ID = scalarValue(strings.TrimPrefix(trimmed, "- id:"))
			continue
		}
		if strings.HasPrefix(trimmed, "path:") {
			current.Path = scalarValue(strings.TrimPrefix(trimmed, "path:"))
			continue
		}
		if strings.HasPrefix(trimmed, "domain:") {
			current.Domain = scalarValue(strings.TrimPrefix(trimmed, "domain:"))
			continue
		}
	}

	if err := flush(); err != nil {
		return nil, err
	}
	if !seenSection {
		return nil, fmt.Errorf("section %q not found", section)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("section %q has no items", section)
	}
	return items, nil
}

func scalarValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "\"'")
	return strings.TrimSpace(value)
}
