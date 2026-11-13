package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	adaptersconfig "agenticlens/internal/adapters/config"
	"gopkg.in/yaml.v3"
)

// workspaceCLICommand declares the fixture's search scope in shared policy.
// It never changes repository fixtures or restores the retired root flag.
func workspaceCLICommand(t *testing.T, searchRoot string, args ...string) *exec.Cmd {
	t.Helper()
	absolute, err := filepath.Abs(searchRoot)
	if err != nil {
		t.Fatal(err)
	}
	project := ""
	if strings.HasPrefix(absolute, filepath.Clean(os.TempDir())+string(filepath.Separator)) {
		if candidate, err := adaptersconfig.FindConfigFile(absolute); err == nil && candidate != "" && strings.HasPrefix(candidate, filepath.Clean(os.TempDir())+string(filepath.Separator)) {
			project = filepath.Dir(candidate)
		}
	}
	if project == "" {
		project = t.TempDir()
		writeCLISharedPolicy(t, project)
	}
	configPath, err := adaptersconfig.FindConfigFile(project)
	if err != nil || configPath == "" {
		t.Fatalf("fixture policy: %s %v", project, err)
	}
	body, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	var document yaml.Node
	if err = yaml.Unmarshal(body, &document); err != nil {
		t.Fatal(err)
	}
	node := document.Content[0]
	workspaces := fixtureMapping(node, "workspaces")

	roots := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	fixtureSet(root, "path", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: absolute})
	discovery := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	fixtureSet(discovery, "max_depth", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "1"})
	fixtureSet(discovery, "max_directories", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "300"})
	fixtureSet(root, "discovery", discovery)
	roots.Content = append(roots.Content, root)
	fixtureSet(workspaces, "roots", roots)
	data, err := yaml.Marshal(&document)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	localPath := filepath.Join(project, ".agenticlens-config.local.yaml")
	local := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if body, err := os.ReadFile(localPath); err == nil {
		var existing yaml.Node
		if err = yaml.Unmarshal(body, &existing); err != nil {
			t.Fatal(err)
		}
		local = existing.Content[0]
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	state := fixtureMapping(local, "agenticlens")
	fixtureSet(state, "state_dir", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: filepath.Join(project, "publication-state")})
	data, err = yaml.Marshal(local)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(localPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(cliBinaryPath, args...)
	command.Dir = project
	return command
}
func fixtureMapping(parent *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			return parent.Content[i+1]
		}
	}
	result := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	fixtureSet(parent, key, result)
	return result
}
func fixtureSet(parent *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(parent.Content); i += 2 {
		if parent.Content[i].Value == key {
			parent.Content[i+1] = value
			return
		}
	}
	parent.Content = append(parent.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func TestCLIRetiredRootFlagRejected(t *testing.T) {
	for _, prefix := range [][]string{nil, {"server"}, {"backfill"}, {"migrate"}, {"format"}} {
		t.Run(strings.Join(prefix, "_")+"_root_override", func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "must-not-create")
			args := append(append([]string{}, prefix...), "--workspaces", destination)
			command := exec.Command(cliBinaryPath, args...)
			output, err := command.CombinedOutput()
			if err == nil || !strings.Contains(string(output), "flag provided but not defined") {
				t.Fatalf("retired root override accepted: %v %s", err, output)
			}
			if _, err = os.Stat(destination); !os.IsNotExist(err) {
				t.Fatalf("flag rejection touched destination: %v", err)
			}
		})
	}
}

func TestCLIConfiguredDepthAndRecordedIdentityWithoutPrivateState(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	command := workspaceCLICommand(t, workspaces, "--list")
	project := command.Dir
	renamed := filepath.Join(workspaces, "arbitrary-label")
	if err := os.Rename(filepath.Join(workspaces, "TICKET-1"), renamed); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(workspaces, "TICKET-99"), 0700); err != nil {
		t.Fatal(err)
	}
	local := filepath.Join(project, ".agenticlens-config.local.yaml")
	if err := os.Remove(local); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, ".agenticlens-config.yaml")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	zero := strings.Replace(string(body), "max_depth: 1", "max_depth: 0", 1)
	if err = os.WriteFile(path, []byte(zero), 0600); err != nil {
		t.Fatal(err)
	}
	rootOnly := exec.Command(cliBinaryPath, "--list")
	rootOnly.Dir = project
	if output, err := rootOnly.CombinedOutput(); err == nil || strings.Contains(string(output), "First Feature") {
		t.Fatalf("root-only scope discovered child ledger: %v %s", err, output)
	}
	if err = os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
	output, err := command.CombinedOutput()
	if err != nil || !strings.Contains(string(output), "TICKET-1") || !strings.Contains(string(output), "TICKET-2") || strings.Contains(string(output), "TICKET-99") {
		t.Fatalf("ledger identity/scope lost: %v %s", err, output)
	}
	for _, path := range []string{local, filepath.Join(project, "publication-state")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("read created private state at %s: %v", path, err)
		}
	}
}

func TestCLIDuplicateTicketRequiresLocationSelection(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	configured := workspaceCLICommand(t, workspaces, "--list")
	duplicate := filepath.Join(workspaces, "copied-location")
	if err := os.Mkdir(duplicate, 0700); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(workspaces, "TICKET-1", "agentic_metrics.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(duplicate, "agentic_metrics.json"), body, 0600); err != nil {
		t.Fatal(err)
	}
	ambiguous := exec.Command(cliBinaryPath, "--ticket-id", "TICKET-1")
	ambiguous.Dir = configured.Dir
	if output, err := ambiguous.CombinedOutput(); err == nil || !strings.Contains(string(output), "ambiguous") {
		t.Fatalf("duplicate ticket silently selected: %v %s", err, output)
	}
	selected := exec.Command(cliBinaryPath, "--ticket-id", "TICKET-1", "--workspace", duplicate)
	selected.Dir = configured.Dir
	if output, err := selected.CombinedOutput(); err != nil || !strings.Contains(string(output), "$1.75") {
		t.Fatalf("explicit ledger directory failed: %v %s", err, output)
	}
	conflict := exec.Command(cliBinaryPath, "--ticket-id", "TICKET-2", "--workspace", duplicate)
	conflict.Dir = configured.Dir
	if output, err := conflict.CombinedOutput(); err == nil || !strings.Contains(string(output), "identity_conflict") {
		t.Fatalf("selection changed ticket identity: %v %s", err, output)
	}
}

// Source extraction fixtures explicitly select their already-created output
// directory; this keeps source/error assertions past unrelated scope admission.
func prepareScopedGenerateFixture(t *testing.T, command *exec.Cmd) {
	t.Helper()
	prepareGenerateFixture(t, command)
	output := ""
	for i, arg := range command.Args {
		if arg == "--json-output" && i+1 < len(command.Args) {
			output = command.Args[i+1]
		}
	}
	if output == "" {
		output = filepath.Join(command.Dir, "agentic_metrics.json")
		command.Args = append(command.Args, "--json-output", output)
	}
	if !filepath.IsAbs(output) {
		output = filepath.Join(command.Dir, output)
	}
	directory := filepath.Dir(output)
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(command.Dir, ".agenticlens-config.yaml")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	var workspaces *yaml.Node
	for i := 0; i < len(doc.Content[0].Content); i += 2 {
		if doc.Content[0].Content[i].Value == "workspaces" {
			workspaces = doc.Content[0].Content[i+1]
		}
	}
	if workspaces == nil {
		return
	}
	var hasRoots bool
	for i := 0; i < len(workspaces.Content); i += 2 {
		if workspaces.Content[i].Value == "roots" {
			hasRoots = true
		}
	}
	if !hasRoots {
		return
	}
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	fixtureSet(root, "path", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: directory})
	discovery := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	fixtureSet(discovery, "max_depth", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "0"})
	fixtureSet(discovery, "max_directories", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "300"})
	fixtureSet(root, "discovery", discovery)
	fixtureSet(workspaces, "roots", &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: []*yaml.Node{root}})
	body, err = yaml.Marshal(&doc)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, body, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCLIComputedOutputsPreserveFactLedgerAndContainment(t *testing.T) {
	workspaces := setupTestWorkspace(t)
	home := filepath.Join(workspaces, "TICKET-1")
	ledger := filepath.Join(home, "agentic_metrics.json")
	original, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	summary := filepath.Join(home, "delivery.md")
	computed := filepath.Join(home, "computed.json")
	cmd := workspaceCLICommand(t, workspaces, "--ticket-id", "TICKET-1", "--append", summary, "--json-output", computed)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("selected output set failed: %v %s", err, output)
	}
	for _, path := range []string{summary, computed} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("output %s missing: %v", path, err)
		}
	}
	after, err := os.ReadFile(ledger)
	if err != nil || string(after) != string(original) {
		t.Fatalf("derived output changed source fact ledger: %v", err)
	}
	outside := filepath.Join(t.TempDir(), "outside.md")
	if err = os.WriteFile(outside, []byte("keep outside"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"--append", outside}, {"--json-output", ledger}} {
		rejected := workspaceCLICommand(t, workspaces, append([]string{"--ticket-id", "TICKET-1"}, args...)...)
		if output, err := rejected.CombinedOutput(); err == nil {
			t.Fatalf("unsafe output accepted: %v %s", args, output)
		}
		if body, err := os.ReadFile(ledger); err != nil || string(body) != string(original) {
			t.Fatalf("rejected output changed source facts: %v", err)
		}
		if body, err := os.ReadFile(outside); err != nil || string(body) != "keep outside" {
			t.Fatalf("rejected output changed outside file: %v", err)
		}
	}
	if _, err := os.Stat(ledger + ".publication-pending"); !os.IsNotExist(err) {
		t.Fatalf("completed/rejected output retained pending marker: %v", err)
	}
}

func TestCLIGenerateRejectsNoncanonicalLedgerBeforeWrites(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(map[bool]string{false: "first-publication", true: "existing-ledger"}[existing], func(t *testing.T) {
			project := t.TempDir()
			canonical := filepath.Join(project, "agentic_metrics.json")
			output := filepath.Join(project, "computed.json")
			summary := filepath.Join(project, "delivery.md")
			accepted := []byte(sampleFactJSON("TEST-91", "Portable existing facts", 60, 0.5))
			if existing {
				if err := os.WriteFile(canonical, accepted, 0600); err != nil {
					t.Fatal(err)
				}
			}
			for path, body := range map[string][]byte{output: accepted, summary: []byte("preserve human evidence")} {
				if err := os.WriteFile(path, body, 0600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(cliBinaryPath, "generate", "--source", "codex", "--codex-home", setupCLIEmptyCodexRoot(t), "--session-id", "fixture-empty-root", "--ticket-id", "TEST-91", "--ticket-title", "Canonical ledger fixture", "--ticket-type", "test", "--json-output", output, "--append", summary)
			cmd.Dir = project
			prepareScopedGenerateFixture(t, cmd)
			result, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(result), "canonical agentic_metrics.json") {
				t.Fatalf("noncanonical generation accepted or wrong boundary reached: %v %s", err, result)
			}
			for path, want := range map[string]string{output: string(accepted), summary: "preserve human evidence"} {
				if got, err := os.ReadFile(path); err != nil || string(got) != want {
					t.Fatalf("rejected generation changed %s: %q %v", path, got, err)
				}
			}
			if existing {
				if got, err := os.ReadFile(canonical); err != nil || string(got) != string(accepted) {
					t.Fatalf("canonical facts changed: %q %v", got, err)
				}
			} else if _, err := os.Stat(canonical); !os.IsNotExist(err) {
				t.Fatalf("rejection created canonical ledger: %v", err)
			}
			for _, path := range []string{filepath.Join(project, ".publication-state"), canonical + ".publication-pending", output + ".publication-pending"} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("rejection started publication at %s: %v", path, err)
				}
			}
		})
	}
}
