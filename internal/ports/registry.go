package ports

import "context"

type BQARegistryReader interface {
	LoadBQARegistry(ctx context.Context) (BQARegistry, error)
}

type BQARegistry struct {
	Agents    []BQARegistryItem
	Skills    []BQARegistryItem
	Workflows []BQARegistryItem
}

type BQARegistryItem struct {
	ID     string
	Path   string
	Domain string
}
