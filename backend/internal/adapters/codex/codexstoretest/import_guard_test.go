package codexstoretest

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// FIX-21 TEST-06: this package imports testing and embeds time zone data, so
// no production package may depend on it. Test builds are not listed by
// go list -deps, so only a production import fails this guard. With -e,
// go list still lists every package's imports when a package has an error,
// such as the web adapter's embed of the gitignored frontend build, so the
// guard does not need that build (R2-T7).
func TestNoProductionPackageImportsCodexStoreTest(t *testing.T) {
	gomod, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		t.Fatalf("go env GOMOD: %v", err)
	}
	root := filepath.Dir(strings.TrimSpace(string(gomod)))
	self, err := exec.Command("go", "list", ".").Output()
	if err != nil {
		t.Fatalf("go list .: %v", err)
	}
	guarded := strings.TrimSpace(string(self))
	list := exec.Command("go", "list", "-e", "-f", `{{.ImportPath}} {{join .Deps " "}}`, "./...")
	list.Dir = root
	out, err := list.Output()
	if err != nil {
		t.Fatalf("go list ./...: %v", err)
	}
	packages := 0
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		fields := strings.Fields(line)
		packages++
		if fields[0] == guarded {
			continue
		}
		for _, dependency := range fields[1:] {
			if dependency == guarded {
				t.Errorf("production package %s depends on %s", fields[0], guarded)
			}
		}
	}
	if packages < 10 {
		t.Fatalf("go list found only %d packages under %s", packages, root)
	}
}
