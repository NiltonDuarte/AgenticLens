package config

import (
	"agenticlens/internal/domain"
	"fmt"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"strings"
)

func ValidateWorkspacePolicy(policy domain.WorkspacesConfig) error {
	if len(policy.Roots) == 0 {
		return fmt.Errorf("workspaces.roots must be an explicit nonempty list; update retired workspaces.dir configuration")
	}
	for i, root := range policy.Roots {
		if strings.TrimSpace(root.Path) == "" || strings.ContainsRune(root.Path, 0) {
			return fmt.Errorf("workspaces.roots[%d].path must be nonblank", i)
		}
		if root.Discovery.MaxDepth < 0 || root.Discovery.MaxDirectories < 1 {
			return fmt.Errorf("workspaces.roots[%d].discovery requires nonnegative max_depth and positive max_directories", i)
		}
		for _, a := range root.Artifacts {
			if !safeArtifact(a.Path) {
				return fmt.Errorf("workspaces.roots[%d].artifacts contains unsafe path %q", i, a.Path)
			}
		}
	}
	return nil
}
func safeArtifact(path string) bool {
	if path == "" || filepath.IsAbs(path) || strings.ContainsAny(path, "\\\x00") {
		return false
	}
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}
func nodeFields(n *yaml.Node, label string, allowed ...string) (map[string]*yaml.Node, error) {
	if err := requireNodeType(n, label, yaml.MappingNode, ""); err != nil {
		return nil, err
	}
	fields := map[string]*yaml.Node{}
	for i := 0; i < len(n.Content); i += 2 {
		key := n.Content[i].Value
		valid := false
		for _, name := range allowed {
			if key == name {
				valid = true
			}
		}
		if !valid || fields[key] != nil {
			return nil, fmt.Errorf("%s contains retired, duplicate or unsupported field %q; update configuration", label, key)
		}
		fields[key] = n.Content[i+1]
	}
	return fields, nil
}
func validateWorkspaceNodes(n *yaml.Node) error {
	if n == nil {
		return nil
	} // Candidate scaffolds may be incomplete; runtime validation requires roots.
	fields, err := nodeFields(n, "workspaces", "roots", "artifacts")
	if err != nil {
		return err
	}
	roots := fields["roots"]
	if roots == nil {
		return fmt.Errorf("workspaces.roots must be explicitly configured")
	}
	if err = requireNodeType(roots, "workspaces.roots", yaml.SequenceNode, ""); err != nil {
		return err
	}
	for i, n := range roots.Content {
		label := fmt.Sprintf("workspaces.roots[%d]", i)
		fields, err := nodeFields(n, label, "path", "discovery", "artifacts")
		if err != nil {
			return err
		}
		if err = requireNodeType(fields["path"], label+".path", yaml.ScalarNode, "!!str"); err != nil {
			return err
		}
		d, err := nodeFields(fields["discovery"], label+".discovery", "max_depth", "max_directories")
		if err != nil {
			return err
		}
		for _, key := range []string{"max_depth", "max_directories"} {
			if err = requireNodeType(d[key], label+".discovery."+key, yaml.ScalarNode, "!!int"); err != nil {
				return err
			}
		}
		if a := fields["artifacts"]; a != nil {
			if err = requireNodeType(a, label+".artifacts", yaml.SequenceNode, ""); err != nil {
				return err
			}
			for _, item := range a.Content {
				if _, err = nodeFields(item, label+".artifacts", "path", "title", "description"); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
