package capture

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// For an existing config, exchange the old and new directory entries atomically
// and verify the actual displaced file. Create new configs only at vacant paths.
// Retain both inodes until commit for recovery across process interruptions.
func applyCodexConfig(stage string, c fileChange, displaced func() error) error {
	if err := os.Mkdir(stage, 0700); err != nil {
		return err
	}
	if !c.remove {
		if err := apply(fileChange{path: filepath.Join(stage, "next"), content: c.content, mode: c.mode}); err != nil {
			return err
		}
	}
	if c.existed {
		prior := filepath.Join(stage, "prior")
		if c.remove {
			if err := moveConfigIfVacant(c.path, prior); err != nil {
				return err
			}
		} else {
			if err := os.Link(filepath.Join(stage, "next"), prior); err != nil {
				return err
			}
			if err := swapConfigEntries(c.path, prior); err != nil {
				return err
			}
		}
		if err := checkConfigImage(prior, c.before, c.mode); err != nil {
			return err
		}
	}
	if err := displaced(); err != nil {
		return err
	}
	if c.remove || c.existed {
		return nil
	}
	return os.Link(filepath.Join(stage, "next"), c.path)
}

func checkConfigImage(path string, expected []byte, mode os.FileMode) error {
	data, actualMode, exists, err := readCodexConfigFile(path)
	if err != nil {
		return err
	}
	if !exists || actualMode != mode || !bytes.Equal(data, expected) {
		return fmt.Errorf("Codex settings changed during installation; retained at %s", path)
	}
	return nil
}

func verifyCodexConfig(stage string, c fileChange) error {
	if c.existed {
		if err := checkConfigImage(filepath.Join(stage, "prior"), c.before, c.mode); err != nil {
			return err
		}
	}
	if c.remove {
		if _, err := os.Lstat(c.path); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return errors.New("Codex settings re-created during uninstallation; preserved")
	}
	mode := c.mode
	if !c.existed {
		mode = 0600
	}
	if err := checkConfigImage(c.path, c.content, mode); err != nil {
		return err
	}
	current, err := os.Lstat(c.path)
	if err != nil {
		return err
	}
	proposed, err := os.Lstat(filepath.Join(stage, "next"))
	if err != nil {
		return err
	}
	if !os.SameFile(current, proposed) {
		return errors.New("Codex configuration replaced during installation; preserved")
	}
	return nil
}

// A prior file was actually displaced only when it exists in this transaction's
// staging area. Never delete or roll back an intervening editor's version.
func rollbackCodexConfig(stage string, c userInstallChange) error {
	info, err := os.Lstat(stage)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm() != 0700 {
		return fmt.Errorf("invalid Codex settings staging directory: %s", stage)
	}
	prior := filepath.Join(stage, "prior")
	next := filepath.Join(stage, "next")
	undo := filepath.Join(stage, "undo")
	previous, err := os.Lstat(prior)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	proposed, err := os.Lstat(next)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	displaced, err := os.Lstat(undo)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	current, err := os.Lstat(c.Path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if proposed != nil && previous != nil && os.SameFile(previous, proposed) {
		// The process stopped between preparing the two links and exchanging.
		return cleanupCodexConfig(stage)
	}
	if previous != nil && current == nil && proposed != nil && displaced == nil {
		return fmt.Errorf("Codex settings removed concurrently; previous version preserved at %s", prior)
	}
	if displaced == nil && proposed != nil && current != nil && os.SameFile(current, proposed) {
		if err := os.Rename(c.Path, undo); err != nil {
			return err
		}
		displaced, err = os.Lstat(undo)
		if err != nil {
			return err
		}
	}
	if displaced != nil {
		mode := c.Mode
		if !c.Existed {
			mode = 0600
		}
		if proposed == nil || !os.SameFile(displaced, proposed) || checkConfigImage(undo, c.Content, mode) != nil {
			// The editor won the race at withdrawal, or updated our inode.
			if err := restoreConfigEntry(undo, c.Path); err != nil {
				return err
			}
			return fmt.Errorf("Codex settings edited during recovery; preserved with prior version in %s", stage)
		}
	}
	if previous != nil {
		if err := restoreConfigEntry(prior, c.Path); err != nil {
			return err
		}
	}
	return cleanupCodexConfig(stage)
}

func restoreConfigEntry(from, to string) error {
	err := moveConfigIfVacant(from, to)
	if errors.Is(err, os.ErrExist) {
		src, srcErr := os.Lstat(from)
		dst, dstErr := os.Lstat(to)
		if srcErr == nil && dstErr == nil && os.SameFile(src, dst) {
			return nil
		}
		return fmt.Errorf("Codex settings changed concurrently at %s; separate version preserved at %s", to, from)
	}
	return err
}

func cleanupCodexConfig(stage string) error {
	if err := cleanupConfigImages(stage); err != nil {
		return err
	}
	return os.Remove(stage)
}

func cleanupConfigImages(stage string) error {
	for _, name := range []string{"undo", "next", "prior"} {
		if err := os.Remove(filepath.Join(stage, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return nil
}
