package backfill

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
)

// discoverWorkspaceLedgers keeps root failures fatal and isolates descendant
// failures. Diagnostics describe uninspected coverage without fabricating a
// ledger, ticket identity or successful empty result for that path.
func discoverWorkspaceLedgers(root string) (paths []string, diagnostics []string, err error) {
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if path == root {
				return walkErr
			}
			diagnostics = append(diagnostics, fmt.Sprintf("unable to inspect workspace path %q: %v; discovery coverage is incomplete", path, walkErr))
			return nil
		}
		if !entry.IsDir() && entry.Name() == "agentic_metrics.json" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, diagnostics, fmt.Errorf("walking workspaces directory %s: %w", root, err)
	}
	sort.Strings(paths)
	return paths, diagnostics, nil
}
