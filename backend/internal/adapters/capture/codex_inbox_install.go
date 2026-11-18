package capture

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"agenticlens/internal/adapters/config"

	"golang.org/x/sys/unix"
)

const codexInboxName = config.CodexInboxName

// CodexBindingInbox resolves the private home before joining the sole directory
// the installer grants. Never resolve a symlink at the writable leaf.
func CodexBindingInbox(home string) (string, error) {
	if !filepath.IsAbs(home) || !utf8.ValidString(home) {
		return "", errors.New("Codex binding inbox requires an absolute Codex home")
	}
	parent, err := absolute(home)
	if err != nil {
		return "", err
	}
	path := filepath.Join(parent, codexInboxName)
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return path, nil
	}
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return "", errors.New("Codex binding inbox must be a private, non-symlink directory")
	}
	return path, nil
}

// CodexInboxInstalled checks the private installer receipt before offering the
// inbox protocol. Older hook installations have no inbox capability.
func CodexInboxInstalled(home string) (bool, error) {
	path, err := CodexBindingInbox(home)
	if err != nil {
		return false, err
	}
	installed, err := codexInboxInReceipt(filepath.Dir(path))
	if err != nil || !installed {
		return false, err
	}
	_, err = os.Lstat(path)
	return err == nil, err
}

func codexInboxInReceipt(home string) (bool, error) {
	content, _, exists, err := readCodexConfigFile(filepath.Join(home, "agenticlens-user-installation.json"))
	if err != nil || !exists {
		return false, err
	}
	var receipt userInstallation
	if err := json.Unmarshal(content, &receipt); err != nil {
		return false, err
	}
	if receipt.Protocol != userCaptureOwnershipProtocol {
		return false, errors.New("unsupported user capture ownership protocol")
	}
	return receipt.BindingInbox && len(receipt.Groups) > 0, nil
}

func CheckCodexInboxSeparation(project, home, inbox string) error {
	paths := []string{filepath.Join(home, "sessions"), filepath.Join(home, "archived_sessions")}
	if project != "" {
		configured, err := config.BindingStatePaths(project)
		if err != nil {
			return err
		}
		// An installation can precede project setup. Where a shared pin exists,
		// resolve its actual development or cached output without running it.
		shared, err := config.FindConfigFile(project)
		if err != nil {
			return err
		}
		if shared != "" {
			entry, err := config.ResolveEntrypoint(project)
			if err != nil {
				return err
			}
			native, err := entry.NativeCandidatePath()
			if err != nil {
				return err
			}
			if _, err := config.OutsideCodexInbox(native, home); err != nil {
				return err
			}
		}
		paths = append(paths, configured...)
	}
	for _, private := range paths {
		physical, err := absolute(private)
		if err != nil {
			return err
		}
		under := func(path, dir string) bool {
			return path == dir || strings.HasPrefix(path, strings.TrimRight(dir, string(filepath.Separator))+string(filepath.Separator))
		}
		if under(inbox, physical) || under(physical, inbox) {
			return fmt.Errorf("Codex binding inbox overlaps private state or source storage: %s", private)
		}
	}
	return nil
}

// Refuse aliases, special files and oversized input before reading user settings.
func readCodexConfigFile(path string) ([]byte, os.FileMode, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, err
	}
	if !info.Mode().IsRegular() {
		return nil, 0, false, fmt.Errorf("Codex configuration must be a regular file: %s", path)
	}
	f, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, 0, false, err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return nil, 0, false, err
	}
	if !os.SameFile(info, opened) || !opened.Mode().IsRegular() || opened.Size() > installJournalLimit {
		return nil, 0, false, errors.New("Codex configuration replaced, oversized, or not regular; preserved")
	}
	data, err := io.ReadAll(io.LimitReader(f, installJournalLimit+1))
	if int64(len(data)) > installJournalLimit {
		return nil, 0, false, errors.New("Codex configuration exceeds transaction size limit")
	}
	return data, opened.Mode().Perm(), true, err
}

func planCodexInbox(home, project string, remove bool, prior userInstallation, next *userInstallation) ([]fileChange, error) {
	if remove && !prior.OwnsInboxGrant {
		return nil, nil
	}
	inbox, err := CodexBindingInbox(home)
	if err != nil {
		return nil, err
	}
	var changes []fileChange
	if !remove {
		if err := CheckCodexInboxSeparation(project, home, inbox); err != nil {
			return nil, err
		}
		if _, err := os.Lstat(inbox); errors.Is(err, os.ErrNotExist) {
			changes = append(changes, fileChange{path: inbox, mkdir: true})
		} else if err != nil {
			return nil, err
		}
		next.BindingInbox = true
	}
	path := filepath.Join(home, "config.toml")
	data, mode, exists, err := readCodexConfigFile(path)
	if err != nil {
		return nil, err
	}
	after, changed, err := editWritableRoot(data, inbox, remove)
	if err != nil {
		return nil, fmt.Errorf("Codex config.toml writable roots: %w", err)
	}
	next.OwnsInboxGrant = !remove && (prior.OwnsInboxGrant || changed)
	next.CreatedInboxConfig = !remove && (prior.CreatedInboxConfig || !exists)
	if changed {
		changes = append(changes, fileChange{path: path, before: data, content: after, existed: exists, mode: mode, remove: remove && prior.CreatedInboxConfig && string(data) == initialInboxConfig(inbox)})
	}
	return changes, nil
}
