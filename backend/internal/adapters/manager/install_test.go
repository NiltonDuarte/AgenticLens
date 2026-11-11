package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"agenticlens/internal/domain"
)

func managerWrite(t *testing.T, path string, body []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, body, mode); err != nil {
		t.Fatal(err)
	}
}

func managerRead(t *testing.T, path string) []byte {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func managerSnapshot(t *testing.T, target string) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, path := range []string{target, target + ".installation.json", target + ".sha256"} {
		body, exists, err := readRegular(path, artifactLimit)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			result[path] = "absent"
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		result[path] = fmt.Sprintf("%o:%s", info.Mode().Perm(), digest(body))
	}
	return result
}

func TestManagerTransactionRecoversActualDurablePrefixes(t *testing.T) {
	for _, initial := range []string{"fresh", "update"} {
		for _, prefix := range []string{"journaled", "manager", "receipt", "checksum", "before-commit"} {
			t.Run(initial+"/"+prefix, func(t *testing.T) {
				root := t.TempDir()
				target := filepath.Join(root, "agentic-lens")
				candidate := filepath.Join(root, "candidate")
				managerWrite(t, candidate, []byte("admitted manager A"), 0700)
				if initial == "update" {
					if err := replaceManager(target, candidate, domain.BuildInfo{Version: "v1.0.0"}); err != nil {
						t.Fatal(err)
					}
				}
				before := managerSnapshot(t, target)
				managerWrite(t, candidate, []byte("admitted manager B"), 0700)
				stop := errors.New("interrupted actual installation")
				err := replaceManagerWithCheckpoint(target, candidate, domain.BuildInfo{Version: "v2.0.0"}, func(stage managerInstallStage, index int) error {
					if prefix == "journaled" && stage == managerInstallJournaled {
						return stop
					}
					if prefix == "before-commit" && stage == managerInstallBeforeCommit {
						return stop
					}
					if stage == managerInstallApplied && map[string]int{"manager": 0, "receipt": 1, "checksum": 2}[prefix] == index && prefix != "journaled" && prefix != "before-commit" {
						return stop
					}
					return nil
				})
				if !errors.Is(err, stop) {
					t.Fatalf("did not reach actual prefix: %v", err)
				}
				var journal transaction
				if err := json.Unmarshal(managerRead(t, target+".pending.json"), &journal); err != nil {
					t.Fatal(err)
				}
				if journal.Target != target || len(journal.Changes) != 3 {
					t.Fatalf("invalid real journal: %+v", journal)
				}
				if prefix != "journaled" && reflect.DeepEqual(before, managerSnapshot(t, target)) {
					t.Fatal("prefix applied no real change")
				}
				if err := recoverManager(target); err != nil {
					t.Fatal(err)
				}
				if got := managerSnapshot(t, target); !reflect.DeepEqual(got, before) {
					t.Fatalf("recovery lost before image: got=%v want=%v", got, before)
				}
				if _, err := os.Stat(target + ".pending.json"); !os.IsNotExist(err) {
					t.Fatalf("recovery retained journal: %v", err)
				}
				if err := replaceManager(target, candidate, domain.BuildInfo{Version: "v2.0.0"}); err != nil {
					t.Fatal(err)
				}
				var installed installation
				if err := json.Unmarshal(managerRead(t, target+".installation.json"), &installed); err != nil {
					t.Fatal(err)
				}
				if installed.Version != "v2.0.0" || installed.SHA256 != digest(managerRead(t, target)) || strings.TrimSpace(string(managerRead(t, target+".sha256"))) != installed.SHA256 {
					t.Fatalf("completed executable/receipt/checksum disagree: %+v", installed)
				}
			})
		}
	}
}

func TestManagerRecoveryPreservesEditsAndRejectsEscapingJournal(t *testing.T) {
	for _, conflict := range []string{"edited executable", "escaping journal"} {
		t.Run(conflict, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "agentic-lens")
			candidate := filepath.Join(root, "candidate")
			managerWrite(t, candidate, []byte("manager A"), 0700)
			if err := replaceManager(target, candidate, domain.BuildInfo{Version: "v1.0.0"}); err != nil {
				t.Fatal(err)
			}
			managerWrite(t, candidate, []byte("manager B"), 0700)
			stop := errors.New("crash after actual executable replace")
			if err := replaceManagerWithCheckpoint(target, candidate, domain.BuildInfo{Version: "v2.0.0"}, func(stage managerInstallStage, index int) error {
				if stage == managerInstallApplied && index == 0 {
					return stop
				}
				return nil
			}); !errors.Is(err, stop) {
				t.Fatalf("prefix: %v", err)
			}
			outside := filepath.Join(root, "outside")
			managerWrite(t, outside, []byte("unrelated owner"), 0600)
			if conflict == "edited executable" {
				managerWrite(t, target, []byte("owner customization"), 0700)
			} else {
				var journal transaction
				if err := json.Unmarshal(managerRead(t, target+".pending.json"), &journal); err != nil {
					t.Fatal(err)
				}
				journal.Changes[0].Path = outside
				raw, err := json.Marshal(journal)
				if err != nil {
					t.Fatal(err)
				}
				managerWrite(t, target+".pending.json", raw, 0600)
			}
			before := managerSnapshot(t, target)
			pending := managerRead(t, target+".pending.json")
			if err := recoverManager(target); err == nil {
				t.Fatal("conflicting recovery succeeded")
			}
			if !reflect.DeepEqual(before, managerSnapshot(t, target)) || !bytes.Equal(pending, managerRead(t, target+".pending.json")) || string(managerRead(t, outside)) != "unrelated owner" {
				t.Fatal("refused recovery changed owner files")
			}
		})
	}
}

func TestManagerReplacementRequiresOwnershipAndExclusiveLease(t *testing.T) {
	for _, scenario := range []string{"unowned", "checksum alone", "orphan checksum", "edited", "edited checksum", "foreign receipt", "busy"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "agentic-lens")
			candidate := filepath.Join(root, "candidate")
			managerWrite(t, candidate, []byte("manager A"), 0700)
			if scenario == "orphan checksum" {
				managerWrite(t, target+".sha256", []byte("unrelated owner checksum\n"), 0600)
			} else if scenario == "unowned" || scenario == "checksum alone" {
				managerWrite(t, target, []byte("owner binary"), 0700)
			} else if err := replaceManager(target, candidate, domain.BuildInfo{Version: "v1.0.0"}); err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "checksum alone":
				managerWrite(t, target+".sha256", []byte(digest(managerRead(t, target))+"\n"), 0600)
			case "edited":
				managerWrite(t, target, []byte("owner edit"), 0700)
			case "edited checksum":
				managerWrite(t, target+".sha256", []byte("custom checksum value\n"), 0600)
			case "foreign receipt":
				var receipt installation
				if err := json.Unmarshal(managerRead(t, target+".installation.json"), &receipt); err != nil {
					t.Fatal(err)
				}
				receipt.Path = filepath.Join(root, "other")
				raw, err := json.Marshal(receipt)
				if err != nil {
					t.Fatal(err)
				}
				managerWrite(t, target+".installation.json", raw, 0600)
			}
			var lease *os.File
			if scenario == "busy" {
				var err error
				lease, err = lockFile(filepath.Join(root, ".agenticlens-manager.lock"))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := lease.Close(); err != nil {
						t.Error(err)
					}
				})
			}
			before := managerSnapshot(t, target)
			managerWrite(t, candidate, []byte("manager B"), 0700)
			if err := replaceManager(target, candidate, domain.BuildInfo{Version: "v2.0.0"}); err == nil {
				t.Fatal("replacement bypassed ownership/lease")
			}
			if !reflect.DeepEqual(before, managerSnapshot(t, target)) {
				t.Fatal("refused replacement changed installation")
			}
			if _, err := os.Stat(target + ".pending.json"); !os.IsNotExist(err) {
				t.Fatalf("ownership refusal mutated transaction journal: %v", err)
			}
		})
	}
}

func TestDownloadVerifiesExactArtifactBeforeAdmission(t *testing.T) {
	name := artifactName("agentic-lens-engine")
	body := []byte("exact release bytes")
	for _, scenario := range []string{"valid", "bin-prefix", "duplicate", "missing", "malformed", "oversized checksum", "checksum mismatch", "checksum HTTP error", "artifact HTTP error"} {
		t.Run(scenario, func(t *testing.T) {
			requests := []string{}
			var requestLock sync.Mutex
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestLock.Lock()
				requests = append(requests, r.URL.Path)
				requestLock.Unlock()
				if strings.HasSuffix(r.URL.Path, "/SHA256SUMS") {
					switch scenario {
					case "checksum HTTP error":
						http.Error(w, "unavailable", http.StatusServiceUnavailable)
					case "duplicate":
						fmt.Fprintf(w, "%s  %s\n%s  bin/%s\n", digest(body), name, digest(body), name)
					case "missing":
						fmt.Fprintf(w, "%s  other\n", digest(body))
					case "malformed":
						fmt.Fprintf(w, "bad  %s\n", name)
					case "oversized checksum":
						if _, err := w.Write(bytes.Repeat([]byte("x"), 1048577)); err != nil {
							t.Errorf("serve oversized checksum: %v", err)
						}
					case "checksum mismatch":
						fmt.Fprintf(w, "%s  %s\n", strings.Repeat("0", 64), name)
					case "bin-prefix":
						fmt.Fprintf(w, "%s  *bin/%s\n", digest(body), name)
					default:
						fmt.Fprintf(w, "%s  %s\n", digest(body), name)
					}
				} else if scenario == "artifact HTTP error" {
					http.Error(w, "unavailable", http.StatusNotFound)
				} else {
					if _, err := w.Write(body); err != nil {
						t.Errorf("serve artifact: %v", err)
					}
				}
			}))
			defer server.Close()
			target := filepath.Join(t.TempDir(), "staged")
			err := downloadVerified(context.Background(), server.Client(), server.URL+"/releases/download/v1.2.3", name, target)
			if scenario == "valid" || scenario == "bin-prefix" {
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(managerRead(t, target), body) {
					t.Fatal("download bytes differ")
				}
			} else if err == nil {
				t.Fatal("invalid release admitted")
			}
			expected := []string{"/releases/download/v1.2.3/SHA256SUMS"}
			if scenario == "valid" || scenario == "bin-prefix" || scenario == "checksum mismatch" || scenario == "artifact HTTP error" {
				expected = append(expected, "/releases/download/v1.2.3/"+name)
			}
			requestLock.Lock()
			defer requestLock.Unlock()
			if !reflect.DeepEqual(requests, expected) {
				t.Fatalf("non-exact or unnecessary download: got=%v want=%v", requests, expected)
			}
		})
	}
}

func TestLocalChecksumFailureDoesNotCreateInstallation(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "candidate")
	destination := filepath.Join(root, "not-created", "agentic-lens")
	managerWrite(t, source, []byte("bad local artifact"), 0700)
	var output bytes.Buffer
	for _, checksum := range []string{"", "not-hex", strings.Repeat("0", 64)} {
		err := Run([]string{"self-update", "--version", "v1.0.0", "--local-bin", source, "--sha256", checksum, "--dest", destination}, strings.NewReader(""), &output, &output)
		if err == nil {
			t.Fatal("invalid checksum admitted")
		}
		if _, err := os.Stat(filepath.Dir(destination)); !os.IsNotExist(err) {
			t.Fatalf("rejected staging created destination: %v", err)
		}
	}
}
