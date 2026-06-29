package doctor

import (
	"context"
	"errors"
	"fmt"

	"github.com/mshegolev/bqa-os/internal/ports"
)

const (
	StatusPass = "PASS"
	StatusFail = "FAIL"
)

type UseCase struct {
	Inspector      ports.WorkspaceInspector
	RegistryReader ports.BQARegistryReader
}

type Result struct {
	Healthy bool
	Checks  []Check
}

type Check struct {
	Name    string
	Path    string
	Status  string
	Message string
}

func (u UseCase) Run(ctx context.Context) (Result, error) {
	if u.Inspector == nil {
		return Result{}, errors.New("workspace inspector is required")
	}
	if u.RegistryReader == nil {
		return Result{}, errors.New("registry reader is required")
	}

	result := Result{Healthy: true}
	result.addCheck(u.directoryCheck(ctx, "workspace", ".bqa"))
	result.addCheck(u.registryCheck(ctx))
	for _, check := range []Check{
		u.directoryCheck(ctx, "memory", ".bqa/memory"),
		u.directoryCheck(ctx, "agents", ".bqa/agents"),
		u.directoryCheck(ctx, "skills", ".bqa/skills"),
		u.directoryCheck(ctx, "workflows", ".bqa/workflows"),
	} {
		result.addCheck(check)
	}
	return result, nil
}

func (u UseCase) registryCheck(ctx context.Context) Check {
	entry, err := u.Inspector.InspectWorkspacePath(ctx, ".bqa/registry")
	if err != nil {
		return fail("registry", ".bqa/registry", err.Error())
	}
	if !entry.Exists {
		return fail("registry", ".bqa/registry", "directory is missing")
	}
	if !entry.IsDir {
		return fail("registry", ".bqa/registry", "path is not a directory")
	}

	registry, err := u.RegistryReader.LoadBQARegistry(ctx)
	if err != nil {
		return fail("registry", ".bqa/registry", err.Error())
	}
	if len(registry.Agents) == 0 {
		return fail("registry", ".bqa/registry", "agents registry is empty")
	}
	if len(registry.Skills) == 0 {
		return fail("registry", ".bqa/registry", "skills registry is empty")
	}
	if len(registry.Workflows) == 0 {
		return fail("registry", ".bqa/registry", "workflows registry is empty")
	}
	return pass(
		"registry",
		".bqa/registry",
		fmt.Sprintf("valid registry with %d agent(s), %d skill(s), %d workflow(s)", len(registry.Agents), len(registry.Skills), len(registry.Workflows)),
	)
}

func (u UseCase) directoryCheck(ctx context.Context, name string, path string) Check {
	entry, err := u.Inspector.InspectWorkspacePath(ctx, path)
	if err != nil {
		return fail(name, path, err.Error())
	}
	if !entry.Exists {
		return fail(name, path, "directory is missing")
	}
	if !entry.IsDir {
		return fail(name, path, "path is not a directory")
	}
	return pass(name, path, "directory exists")
}

func (r *Result) addCheck(check Check) {
	r.Checks = append(r.Checks, check)
	if check.Status == StatusFail {
		r.Healthy = false
	}
}

func pass(name string, path string, message string) Check {
	return Check{Name: name, Path: path, Status: StatusPass, Message: message}
}

func fail(name string, path string, message string) Check {
	return Check{Name: name, Path: path, Status: StatusFail, Message: message}
}
