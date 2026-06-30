package catalog

import "strings"

// renderRegistry emits a registry YAML section for the given entries.
func renderRegistry(key, dir string, entries []Entry) string {
	var b strings.Builder
	b.WriteString(key + ":\n")
	for _, e := range entries {
		b.WriteString("  - id: " + e.ID + "\n")
		b.WriteString("    path: " + dir + "/" + e.ID + ".md\n")
		b.WriteString("    domain: " + e.Domain + "\n")
	}
	return b.String()
}

// registryAgents is the ordered agent list used for registry derivation.
var registryAgents = []Entry{
	{ID: "etl-qa-agent", Domain: "etl"},
	{ID: "runtime-agent", Domain: "runtime"},
}

// registryWorkflows is the curated workflow list exposed in the registry.
var registryWorkflows = []Entry{
	{ID: "etl-verification-workflow", Domain: "etl"},
	{ID: "session-knowledge-workflow", Domain: "memory"},
}

// RegistryAgentsYAML renders the agents registry section.
func RegistryAgentsYAML() string {
	return renderRegistry("agents", "agents", registryAgents)
}

// RegistrySkillsYAML renders the skills registry section.
func RegistrySkillsYAML() string {
	return renderRegistry("skills", "skills", skills)
}

// RegistryWorkflowsYAML renders the workflows registry section.
func RegistryWorkflowsYAML() string {
	return renderRegistry("workflows", "workflows", registryWorkflows)
}
