package manager

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/config"
)

type rejectedPlatformTransport struct{ requests int }

func (transport *rejectedPlatformTransport) RoundTrip(*http.Request) (*http.Response, error) {
	transport.requests++
	return nil, errors.New("unexpected network request for unsupported platform")
}

func TestSupportedPlatformSelections(t *testing.T) {
	for _, platform := range []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64"} {
		t.Run(platform, func(t *testing.T) {
			operatingSystem, architecture, err := parsePlatform(platform)
			if err != nil || operatingSystem+"-"+architecture != platform {
				t.Fatalf("supported target rejected or changed: %q %q %v", operatingSystem, architecture, err)
			}
			if got := platformArtifactName("agentic-lens-engine", operatingSystem, architecture); got != "agentic-lens-engine-"+platform {
				t.Fatalf("supported release name changed: %s", got)
			}
		})
	}
}

func TestUnsupportedPlatformInstallHasNoNetworkOrInstallationWrites(t *testing.T) {
	transport := &rejectedPlatformTransport{}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	t.Cleanup(func() { http.DefaultTransport = previous })
	t.Setenv("AGENTIC_LENS_REPO", "fixture/releases")
	t.Setenv("AGENTIC_LENS_VERSION", "")
	for _, platform := range []string{"windows-amd64", "windows-arm64", "freebsd-amd64", "linux-386", "darwin-386", "linux-amd64-extra"} {
		t.Run(platform, func(t *testing.T) {
			project := t.TempDir()
			cache := filepath.Join(project, "uncreated cache")
			policy := []byte("agenticlens:\n  version: v9.8.7\n")
			config := filepath.Join(project, ".agenticlens-config.yaml")
			managerWrite(t, config, policy, 0600)
			var output bytes.Buffer
			err := Run([]string{"--al-project", project, "install", "v9.8.7", "--platform", platform, "--install-root", cache}, strings.NewReader(""), &output, &output)
			if err == nil || !strings.Contains(err.Error(), "unsupported native platform") {
				t.Fatalf("unsupported selection reached another boundary: %v; %s", err, output.String())
			}
			if transport.requests != 0 {
				t.Fatalf("unsupported selection made %d network requests", transport.requests)
			}
			if _, err := os.Stat(cache); !os.IsNotExist(err) {
				t.Fatalf("unsupported selection created installation root: %v", err)
			}
			if !bytes.Equal(managerRead(t, config), policy) {
				t.Fatal("unsupported selection changed shared policy")
			}
			entries, err := os.ReadDir(project)
			if err != nil || len(entries) != 1 || entries[0].Name() != ".agenticlens-config.yaml" {
				t.Fatalf("unexpected project writes: %v %v", entries, err)
			}
		})
	}
}

func TestWindowsPEArtifactCannotBeAdmitted(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "probe.go")
	artifact := filepath.Join(root, "unsupported-artifact")
	managerWrite(t, source, []byte("package main\nfunc main() {}\n"), 0600)
	// A real PE artifact makes this a platform-removal regression, rather than
	// another malformed-header rejection. The fixture is never executed.
	command := exec.Command("go", "build", "-o", artifact, source)
	command.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build negative PE fixture: %v\n%s", err, output)
	}
	if err := os.Chmod(artifact, 0700); err != nil {
		t.Fatal(err)
	}
	if err := config.ValidateNativePlatform(artifact, "windows", "amd64"); err == nil {
		t.Fatalf("PE artifact reached native admission: %v", err)
	}
}
