package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"agenticlens/internal/domain"
)

type installation struct {
	Protocol int    `json:"protocol"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
	Version  string `json:"version"`
}

type change struct {
	Path    string      `json:"path"`
	Before  []byte      `json:"before"`
	After   []byte      `json:"after"`
	Existed bool        `json:"existed"`
	Mode    fs.FileMode `json:"mode"`
}

type transaction struct {
	Protocol int      `json:"protocol"`
	Target   string   `json:"target"`
	Changes  []change `json:"changes"`
}

type managerInstallStage uint8

const (
	managerInstallJournaled managerInstallStage = iota
	managerInstallApplied
	managerInstallBeforeCommit
)

func replaceManager(target, candidate string, identity domain.BuildInfo) (returnErr error) {
	return replaceManagerWithCheckpoint(target, candidate, identity, nil)
}

func replaceManagerWithCheckpoint(target, candidate string, identity domain.BuildInfo, checkpoint func(managerInstallStage, int) error) (returnErr error) {
	content, _, err := readRegular(candidate, artifactLimit)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return err
	}
	lease, err := lockFile(filepath.Join(filepath.Dir(target), ".agenticlens-manager.lock"))
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lease.Close()) }()
	if err := recoverManager(target); err != nil {
		return err
	}
	before, existed, err := readRegular(target, artifactLimit)
	if err != nil {
		return err
	}
	receiptPath := target + ".installation.json"
	receiptBytes, hasReceipt, err := readRegular(receiptPath, 1048576)
	if err != nil {
		return err
	}
	var owned installation
	if existed {
		if !hasReceipt || json.Unmarshal(receiptBytes, &owned) != nil || owned.Protocol != Protocol || owned.Path != target || owned.SHA256 != digest(before) {
			return fmt.Errorf("installed manager %s is modified or unowned; preserve it and explicitly resolve its installation", target)
		}
	} else if hasReceipt {
		return errors.New("manager ownership record exists without its executable; explicitly resolve the incomplete installation")
	}
	checksumBytes, hasChecksum, err := readRegular(target+".sha256", 1048576)
	if err != nil {
		return err
	}
	if hasChecksum && (!existed || !hasReceipt || !bytes.Equal(checksumBytes, []byte(owned.SHA256+"\n"))) {
		return fmt.Errorf("manager checksum sidecar %s is modified or unowned; preserve it and explicitly resolve its installation", target+".sha256")
	}
	receipt := installation{Protocol: Protocol, Path: target, SHA256: digest(content), Version: identity.Version}
	metadata, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	metadata = append(metadata, '\n')
	values := []struct {
		path string
		body []byte
		mode fs.FileMode
	}{{target, content, 0700}, {receiptPath, metadata, 0600}, {target + ".sha256", []byte(receipt.SHA256 + "\n"), 0600}}
	journal := transaction{Protocol: Protocol, Target: target, Changes: []change{}}
	for _, value := range values {
		previous, exists, err := readRegular(value.path, artifactLimit)
		if err != nil {
			return err
		}
		mode := value.mode
		if exists {
			info, err := os.Stat(value.path)
			if err != nil {
				return err
			}
			mode = info.Mode().Perm()
		}
		journal.Changes = append(journal.Changes, change{Path: value.path, Before: previous, After: value.body, Existed: exists, Mode: mode})
	}
	raw, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	if err := atomicWrite(target+".pending.json", raw, 0600); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint(managerInstallJournaled, -1); err != nil {
			return err
		}
	}
	for index, item := range journal.Changes {
		current, exists, err := readRegular(item.Path, artifactLimit)
		if err != nil {
			return errors.Join(err, recoverManager(target))
		}
		if exists != item.Existed || !bytes.Equal(current, item.Before) {
			return fmt.Errorf("installation target changed during update: %s; pending recovery retained", item.Path)
		}
		mode := fs.FileMode(0600)
		if item.Path == target {
			mode = 0700
		}
		if err := atomicWrite(item.Path, item.After, mode); err != nil {
			return errors.Join(err, recoverManager(target))
		}
		if checkpoint != nil {
			if err := checkpoint(managerInstallApplied, index); err != nil {
				return err
			}
		}
	}
	if checkpoint != nil {
		if err := checkpoint(managerInstallBeforeCommit, -1); err != nil {
			return err
		}
	}
	if err := os.Remove(target + ".pending.json"); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(target))
}

func recoverManager(target string) error {
	path := target + ".pending.json"
	raw, exists, err := readRegular(path, artifactLimit*4)
	if err != nil || !exists {
		return err
	}
	var journal transaction
	if err := json.Unmarshal(raw, &journal); err != nil {
		return fmt.Errorf("pending manager installation: %w", err)
	}
	if journal.Protocol != Protocol || journal.Target != target || len(journal.Changes) != 3 {
		return errors.New("invalid pending manager transaction; preserved")
	}
	allowed := map[string]bool{target: true, target + ".installation.json": true, target + ".sha256": true}
	for _, item := range journal.Changes {
		if !allowed[item.Path] {
			return errors.New("pending manager transaction escapes its installation; preserved")
		}
		delete(allowed, item.Path)
	}
	for i := len(journal.Changes) - 1; i >= 0; i-- {
		item := journal.Changes[i]
		current, exists, err := readRegular(item.Path, artifactLimit)
		if err != nil {
			return err
		}
		if exists == item.Existed && bytes.Equal(current, item.Before) {
			continue
		}
		if !exists || !bytes.Equal(current, item.After) {
			return fmt.Errorf("pending manager target edited; preserved: %s", item.Path)
		}
		if item.Existed {
			if err := atomicWrite(item.Path, item.Before, item.Mode); err != nil {
				return err
			}
		} else if err := os.Remove(item.Path); err != nil {
			return err
		}
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(target))
}

func readRegular(path string, limit int64) ([]byte, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, false, fmt.Errorf("installation file is not regular or exceeds its limit: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	opened, statErr := file.Stat()
	if statErr != nil || !os.SameFile(info, opened) {
		return nil, false, errors.Join(errors.New("installation file changed while opening"), statErr, file.Close())
	}
	body, readErr := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := file.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return nil, false, err
	}
	if int64(len(body)) > limit {
		return nil, false, fmt.Errorf("installation file exceeds limit: %s", path)
	}
	return body, true, nil
}

func atomicWrite(path string, body []byte, mode fs.FileMode) (returnErr error) {
	file, err := os.CreateTemp(filepath.Dir(path), ".agenticlens-install-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	if err := file.Chmod(mode); err != nil {
		return errors.Join(err, file.Close())
	}
	if _, err := file.Write(body); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
