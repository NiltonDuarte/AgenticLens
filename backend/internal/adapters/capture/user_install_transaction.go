package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type userInstallChange struct {
	Path       string `json:"path"`
	Content    []byte `json:"content"`
	Before     []byte `json:"before"`
	Remove     bool   `json:"remove"`
	Existed    bool   `json:"existed"`
	Executable bool   `json:"executable"`
}

type userInstallJournal struct {
	Version int                 `json:"version"`
	Homes   []string            `json:"homes,omitempty"`
	Changes []userInstallChange `json:"changes"`
}

type userInstallStage uint8

const (
	userInstallJournaled userInstallStage = iota
	userInstallApplied
	userInstallBeforeCommit
	userInstallResourcesDiscovered
)

// userJournalHomes reads resources before recovery; callers acquire every home
// lease while holding the root lease before restoring even a single byte.
func pendingUserHomes(directory string) ([]string, error) {
	path := filepath.Join(directory, "agenticlens-user-install.pending.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var journal userInstallJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		return nil, err
	}
	if journal.Version != 1 {
		return nil, errors.New("unsupported pending user installation")
	}
	return userJournalHomes(journal), nil
}

func userJournalHomes(journal userInstallJournal) []string {
	homes := append([]string(nil), journal.Homes...)
	for _, change := range journal.Changes {
		switch filepath.Base(change.Path) {
		case "hooks.json", "agenticlens-user-installation.json", "agenticlens-installation.json", "install-codex.json":
			homes = append(homes, filepath.Dir(change.Path))
		}
	}
	return homes
}

// recoverUserInstallation restores only unchanged transaction-owned bytes. A
// later user edit remains a visible conflict, never an overwrite instruction.
func recoverUserInstallation(home string) error {
	path := filepath.Join(home, "agenticlens-user-install.pending.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var journal userInstallJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		return err
	}
	if journal.Version != 1 {
		return errors.New("unsupported pending user installation")
	}
	if err := rollbackUserJournal(journal); err != nil {
		return fmt.Errorf("user installation recovery incomplete; prior bytes retained: %w", err)
	}
	return os.Remove(path)
}

func rollbackUserJournal(journal userInstallJournal) error {
	for i := len(journal.Changes) - 1; i >= 0; i-- {
		change := journal.Changes[i]
		current, err := os.ReadFile(change.Path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if (err == nil) == change.Existed && bytes.Equal(current, change.Before) {
			continue
		}
		if change.Remove {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("transaction target edited; preserved: %s", change.Path)
			}
		} else if err != nil || !bytes.Equal(current, change.Content) {
			return fmt.Errorf("transaction target edited; preserved: %s", change.Path)
		}
		if err := apply(fileChange{path: change.Path, content: change.Before, remove: !change.Existed}); err != nil {
			return err
		}
		if change.Existed && change.Executable {
			if err := os.Chmod(change.Path, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyUserInstallation(home, dispatcher string, changes []fileChange) error {
	return applyUserInstallationWithCheckpoint(home, dispatcher, changes, nil, nil)
}

func applyUserInstallationWithCheckpoint(home, dispatcher string, changes []fileChange, homes []string, checkpoint func(userInstallStage, int) error) error {
	journal := userInstallJournal{Version: 1, Homes: append([]string(nil), homes...)}
	for _, change := range changes {
		journal.Changes = append(journal.Changes, userInstallChange{Path: change.path, Content: change.content, Before: change.before, Remove: change.remove, Existed: change.existed, Executable: change.path == dispatcher})
	}
	body, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	path := filepath.Join(home, "agenticlens-user-install.pending.json")
	if err := apply(fileChange{path: path, content: body}); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint(userInstallJournaled, -1); err != nil {
			return err
		}
	}
	for index, change := range changes {
		applyErr := apply(change)
		if applyErr == nil && change.path == dispatcher {
			applyErr = os.Chmod(dispatcher, 0755)
		}
		if applyErr != nil {
			rollbackErr := rollbackUserJournal(journal)
			if rollbackErr == nil {
				rollbackErr = os.Remove(path)
			}
			return errors.Join(applyErr, rollbackErr)
		}
		if checkpoint != nil {
			if err := checkpoint(userInstallApplied, index); err != nil {
				return err
			}
		}
	}
	if checkpoint != nil {
		if err := checkpoint(userInstallBeforeCommit, len(changes)); err != nil {
			return err
		}
	}
	return os.Remove(path)
}
