package capture

import (
	config "agenticlens/internal/adapters/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUserHooksInstallIdempotentAndOwnedRemoval(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "codex")
	cache := filepath.Join(root, "native")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	hooksPath := filepath.Join(home, "hooks.json")
	foreign := `{"hooks":{"Stop":[{"hooks":[{"type":"command","command":"foreign command"}]}]},"foreign_setting":true}`
	if err := os.WriteFile(hooksPath, []byte(foreign), 0600); err != nil {
		t.Fatal(err)
	}
	binary, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	opts := UserInstallOptions{Home: home, Entrypoint: binary, DryRun: true}
	if _, err := InstallUserCodex(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if body, err := os.ReadFile(hooksPath); err != nil || string(body) != foreign {
		t.Fatalf("dry-run changed hooks: %s %v", body, err)
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("dry-run created cache: %v", err)
	}
	opts.DryRun = false
	if _, err := InstallUserCodex(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(first, []byte(root+"/project")) || bytes.Contains(first, []byte("--repo")) || bytes.Contains(first, []byte("--store")) {
		t.Fatalf("project policy frozen into user hooks: %s", first)
	}
	var doc struct {
		Hooks   map[string][]json.RawMessage `json:"hooks"`
		Foreign bool                         `json:"foreign_setting"`
	}
	if err := json.Unmarshal(first, &doc); err != nil {
		t.Fatal(err)
	}
	if !doc.Foreign || len(doc.Hooks) != 6 || len(doc.Hooks["Stop"]) != 2 {
		t.Fatalf("additive owned groups incorrect: %+v", doc)
	}
	for event, groups := range doc.Hooks {
		if len(groups) == 0 || !strings.Contains(string(groups[len(groups)-1]), "codex-hook") {
			t.Fatalf("missing native dispatch for %s: %s", event, groups)
		}
	}
	var receipt userInstallation
	if err := json.Unmarshal(mustReadUserInstall(t, filepath.Join(home, "agenticlens-user-installation.json")), &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Protocol != 2 || receipt.Manager != binary {
		t.Fatalf("hook receipt owns the wrong manager path/protocol: %+v", receipt)
	}
	beforeManager, err := os.Stat(binary)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := InstallUserCodex(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	afterManager, err := os.Stat(binary)
	if err != nil || !os.SameFile(beforeManager, afterManager) {
		t.Fatalf("idempotent hook install replaced manager: %v", err)
	}
	if again, err := os.ReadFile(hooksPath); err != nil || !bytes.Equal(first, again) {
		t.Fatalf("repeat duplicated or rewrote hooks: %v", err)
	}
	opts.Remove = true
	if _, err := InstallUserCodex(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(hooksPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	// Decode anew because json.Unmarshal retains absent map keys on reuse.
	var remaining struct {
		Hooks   map[string][]json.RawMessage `json:"hooks"`
		Foreign bool                         `json:"foreign_setting"`
	}
	if err := json.Unmarshal(body, &remaining); err != nil {
		t.Fatal(err)
	}
	if !remaining.Foreign || len(remaining.Hooks) != 1 || len(remaining.Hooks["Stop"]) != 1 || !bytes.Contains(body, []byte("foreign command")) {
		t.Fatalf("removal lost unrelated content: %s", body)
	}
}

func TestUserHooksModifiedOwnedCommandPreserved(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "codex")
	binary, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	opts := UserInstallOptions{Home: home, Entrypoint: binary}
	if _, err := InstallUserCodex(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, "hooks.json")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.ReplaceAll(body, []byte("codex-hook"), []byte("user-edited-hook"))
	if bytes.Equal(body, changed) {
		t.Fatal("fixture did not modify owned command")
	}
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	opts.Remove = true
	if _, err := InstallUserCodex(context.Background(), opts); err == nil {
		t.Fatal("modified owned hook removed")
	}
	if after, err := os.ReadFile(path); err != nil || !bytes.Equal(changed, after) {
		t.Fatalf("rejected removal mutated user content: %v", err)
	}
}

func TestRetainedExecutableSurvivesMutableOutputReplacement(t *testing.T) {
	binary, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	buildA := mustReadUserInstall(t, binary)
	buildB := append(append([]byte{}, buildA...), []byte("AL79 newer completed native output\n")...)
	root := t.TempDir()
	source := filepath.Join(root, "mutable")
	if err := os.WriteFile(source, buildA, 0700); err != nil {
		t.Fatal(err)
	}
	retained, err := RetainExecutable(source, filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, buildB, 0700); err != nil {
		t.Fatal(err)
	}
	next, err := RetainExecutable(source, filepath.Join(root, "state"))
	if err != nil {
		t.Fatal(err)
	}
	if next == retained {
		t.Fatal("changed build reused old artifact")
	}
	if body, err := os.ReadFile(retained); err != nil || !bytes.Equal(body, buildA) {
		t.Fatalf("active artifact changed: %s %v", body, err)
	}
	if body, err := os.ReadFile(next); err != nil || !bytes.Equal(body, buildB) {
		t.Fatalf("new build unavailable: %s %v", body, err)
	}
	same, err := RetainExecutable(source, filepath.Join(root, "state"))
	if err != nil || same != next {
		t.Fatalf("unchanged build not reused: %s %v", same, err)
	}
}

func mustReadUserInstall(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func init() {
	if os.Getenv("AL79_CAPTURE_NATIVE_FIXTURE") == "" {
		return
	}
	switch strings.Join(os.Args[1:], " ") {
	case "engine protocol":
		fmt.Println("1")
	case "capture protocol":
		if os.Getenv("AL79_CAPTURE_NATIVE_FIXTURE") == "old-interface" {
			fmt.Println("1")
		} else {
			fmt.Println(HandlerProtocolVersion)
		}
	case "manager protocol", "dispatch protocol":
		if os.Getenv("AL79_CAPTURE_NATIVE_FIXTURE") == "bad-manager" {
			fmt.Println("0")
		} else {
			fmt.Println("1")
		}
	case "version --short":
		fmt.Println("v9.8.7")
	case "version --json", "--manager-version --json":
		fmt.Println(`{"version":"v9.8.7","commit":"native-test-fixture","build_date":"test"}`)
	default:
		return
	}
	os.Exit(0)
}
func compatibleCaptureTestExecutable(t *testing.T) (string, error) {
	t.Helper()
	t.Setenv("AL79_CAPTURE_NATIVE_FIXTURE", "1")
	t.Setenv("GORACE", strings.TrimSpace(os.Getenv("GORACE")+" atexit_sleep_ms=0"))
	return os.Executable()
}

func TestIncompatibleUserManagerDoesNotCreateOrChangeOwnedState(t *testing.T) {
	for _, owned := range []bool{false, true} {
		t.Run(fmt.Sprint(owned), func(t *testing.T) {
			root := t.TempDir()
			binary, err := compatibleCaptureTestExecutable(t)
			if err != nil {
				t.Fatal(err)
			}
			options := UserInstallOptions{Home: filepath.Join(root, "home"), Entrypoint: binary}
			before := map[string][]byte{}
			if owned {
				if _, err := InstallUserCodex(context.Background(), options); err != nil {
					t.Fatal(err)
				}
				for _, path := range []string{filepath.Join(options.Home, "hooks.json"), filepath.Join(options.Home, "agenticlens-user-installation.json")} {
					before[path] = mustReadUserInstall(t, path)
				}
			}
			t.Setenv("AL79_CAPTURE_NATIVE_FIXTURE", "bad-manager")
			if _, err := InstallUserCodex(context.Background(), options); err == nil {
				t.Fatal("incompatible manager accepted")
			}
			if !owned {
				for _, path := range []string{options.Home, filepath.Join(root, "cache")} {
					if _, err := os.Stat(path); !os.IsNotExist(err) {
						t.Fatalf("invalid manager mutated fresh destination %s: %v", path, err)
					}
				}
			}
			for path, body := range before {
				if !bytes.Equal(body, mustReadUserInstall(t, path)) {
					t.Fatalf("invalid manager changed owned path %s", path)
				}
			}
		})
	}
}
func TestRetainedAdmissionRejectsSourceChangedAfterValidation(t *testing.T) {
	for _, when := range []string{"before-retention", "opened-truncation", "copied-same-size-restored-time"} {
		t.Run(when, func(t *testing.T) {
			binary, err := compatibleCaptureTestExecutable(t)
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			source := filepath.Join(root, "mutable-native")
			original := mustReadUserInstall(t, binary)
			if err := os.WriteFile(source, original, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := config.ValidateNativeExecutable(source); err != nil {
				t.Fatalf("selection's valid native control: %v", err)
			}
			info, err := os.Stat(source)
			if err != nil {
				t.Fatal(err)
			}
			state := filepath.Join(root, "state")
			if err := os.Mkdir(state, 0700); err != nil {
				t.Fatal(err)
			}
			pending := filepath.Join(state, "pending-work")
			if err := os.WriteFile(pending, []byte("unresolved generation retained"), 0600); err != nil {
				t.Fatal(err)
			}
			if when == "before-retention" {
				if err := os.Truncate(source, 0); err != nil {
					t.Fatal(err)
				}
			}
			checkpoint := func(stage retainStage) error {
				if when == "opened-truncation" && stage == retainSourceOpened {
					return os.Truncate(source, info.Size()/2)
				}
				if when == "copied-same-size-restored-time" && stage == retainSourceCopied {
					changed := append([]byte{}, original...)
					changed[len(changed)-1] ^= 1
					if err := os.WriteFile(source, changed, 0700); err != nil {
						return err
					}
					return os.Chtimes(source, info.ModTime(), info.ModTime())
				}
				return nil
			}
			result, err := retainExecutableWithCheckpoint(source, state, checkpoint)
			if err == nil || result != "" {
				t.Fatalf("changed source admitted as complete build: %s %v", result, err)
			}
			if !bytes.Equal(mustReadUserInstall(t, pending), []byte("unresolved generation retained")) {
				t.Fatal("rejected retention changed pending work")
			}
			entries, readErr := os.ReadDir(filepath.Join(state, "native"))
			if readErr != nil && !os.IsNotExist(readErr) {
				t.Fatal(readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("failed admission left retained or temporary executable: %+v", entries)
			}
		})
	}
}

func TestUserHooksDoNotOwnOrReplaceManager(t *testing.T) {
	root := t.TempDir()
	binary, err := compatibleCaptureTestExecutable(t)
	if err != nil {
		t.Fatal(err)
	}
	manager := filepath.Join(root, "manager")
	original := mustReadUserInstall(t, binary)
	if err := os.WriteFile(manager, original, 0700); err != nil {
		t.Fatal(err)
	}
	homeA, homeB := filepath.Join(root, "home-a"), filepath.Join(root, "home-b")
	options := UserInstallOptions{Home: homeA, Entrypoint: manager}
	if _, err := InstallUserCodex(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	hooks := mustReadUserInstall(t, filepath.Join(homeA, "hooks.json"))
	receipt := mustReadUserInstall(t, filepath.Join(homeA, "agenticlens-user-installation.json"))
	upgraded := append(append([]byte{}, original...), []byte("completed manager update")...)
	if err := os.WriteFile(manager, upgraded, 0700); err != nil {
		t.Fatal(err)
	}
	options.Home = homeB
	if _, err := InstallUserCodex(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(upgraded, mustReadUserInstall(t, manager)) {
		t.Fatal("hook installer replaced manager")
	}
	if !bytes.Equal(hooks, mustReadUserInstall(t, filepath.Join(homeA, "hooks.json"))) || !bytes.Equal(receipt, mustReadUserInstall(t, filepath.Join(homeA, "agenticlens-user-installation.json"))) {
		t.Fatal("another home's hook install rewrote first home")
	}
	options.Home, options.Remove = homeA, true
	if _, err := InstallUserCodex(context.Background(), options); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(upgraded, mustReadUserInstall(t, manager)) {
		t.Fatal("hook uninstall removed or replaced manager")
	}
	for _, path := range []string{filepath.Join(root, "cache"), manager + ".sha256", manager + ".installation.json"} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("hook installer acquired manager ownership at %s: %v", path, err)
		}
	}
}
