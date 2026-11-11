package capture_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func historicalCheckHook(t *testing.T, wrapper string) []byte {
	t.Helper()
	// Sanitized output of renderCheckHook at 4e5c4583, before Windows removal.
	// Only the encoded wrapper and its single quoted command argument vary.
	body := checkRead(t, "testdata/git-hook-v1-generated.sh")
	const originalWrapper = "/fixture owner's/bin/agentic-lens"
	originalEncoded := base64.StdEncoding.EncodeToString([]byte(originalWrapper))
	for _, substitution := range [][2]string{
		{originalEncoded, base64.StdEncoding.EncodeToString([]byte(wrapper))},
		{capture.QuoteArgument(originalWrapper), capture.QuoteArgument(filepath.ToSlash(wrapper))},
	} {
		if bytes.Count(body, []byte(substitution[0])) != 1 {
			t.Fatalf("historical fixture has no unique wrapper marker %q", substitution[0])
		}
		body = bytes.Replace(body, []byte(substitution[0]), []byte(substitution[1]), 1)
	}
	return body
}

func TestHistoricalGitCheckCanUpgradeOrRemoveWithoutExecutingLegacyScript(t *testing.T) {
	for _, remove := range []bool{false, true} {
		name := "upgrade"
		if remove {
			name = "remove"
		}
		t.Run(name, func(t *testing.T) {
			repo := checkRepository(t)
			wrapper := filepath.Join(t.TempDir(), "manager owner's $value with spaces")
			hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
			before := historicalCheckHook(t, wrapper)
			checkWrite(t, hook, string(before), 0755)
			options := capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper, Remove: remove, DryRun: true}
			want := "update"
			if remove {
				want = "remove"
			}
			if result := installCheck(t, options); result.Action != want || !bytes.Equal(before, checkRead(t, hook)) {
				t.Fatalf("historical preview changed source or wrong action: %+v", result)
			}
			options.DryRun = false
			if result := installCheck(t, options); result.Action != want {
				t.Fatalf("historical operation failed: %+v", result)
			}
			if remove {
				if _, err := os.Stat(hook); !os.IsNotExist(err) {
					t.Fatalf("historical hook retained: %v", err)
				}
				return
			}
			after := checkRead(t, hook)
			if bytes.Equal(before, after) || bytes.Contains(after, []byte("tree_kill")) || !bytes.Contains(after, []byte(capture.QuoteArgument(filepath.ToSlash(wrapper))+" '--al-offline' '--al-project'")) {
				t.Fatalf("upgrade did not retain literal wrapper in current hook: %s", after)
			}
			if result := installCheck(t, options); result.Action != "unchanged" {
				t.Fatalf("upgraded hook is not stable: %+v", result)
			}
		})
	}
}

func TestHistoricalGitCheckPreservesEveryEditedVariant(t *testing.T) {
	repo := checkRepository(t)
	wrapper := filepath.Join(t.TempDir(), "owner's manager")
	hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
	original := string(historicalCheckHook(t, wrapper))
	encoded := base64.StdEncoding.EncodeToString([]byte(wrapper))
	for name, body := range map[string]string{
		"appended command":        original + "printf 'owner customization'\n",
		"changed deadline":        strings.Replace(original, "sleep 3", "sleep 9", 1),
		"changed wrapper command": strings.Replace(original, capture.QuoteArgument(filepath.ToSlash(wrapper)), "'another executable'", 1),
		"noncanonical header":     strings.Replace(original, encoded+"\n", encoded+"\r\n", 1),
	} {
		t.Run(name, func(t *testing.T) {
			checkWrite(t, hook, body, 0755)
			for _, remove := range []bool{false, true} {
				for _, dryRun := range []bool{false, true} {
					_, err := capture.InstallCheck(context.Background(), capture.CheckInstallOptions{Repository: repo, Wrapper: wrapper, Remove: remove, DryRun: dryRun})
					if err == nil || !strings.Contains(err.Error(), "existing or edited hook") {
						t.Fatalf("edited hook accepted remove=%t dry=%t: %v", remove, dryRun, err)
					}
					if !bytes.Equal([]byte(body), checkRead(t, hook)) {
						t.Fatal("edited historical hook was changed")
					}
				}
			}
		})
	}
}
