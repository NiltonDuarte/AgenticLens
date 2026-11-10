package capture

import (
	config "agenticlens/internal/adapters/config"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// RetainExecutable gives admitted workers immutable bytes for all their children.
// Mutable development output may be rebuilt without replacing the active build.
func RetainExecutable(binary, directory string) (string, error) {
	return retainExecutableWithCheckpoint(binary, directory, nil)
}

type retainStage uint8

const (
	retainSourceOpened retainStage = iota
	retainSourceCopied
)

func retainExecutableWithCheckpoint(binary, directory string, checkpoint func(retainStage) error) (result string, returnErr error) {
	source, err := os.Open(binary)
	if err != nil {
		return "", err
	}
	defer func() { returnErr = errors.Join(returnErr, source.Close()) }()
	info, err := source.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() <= 0 || (runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0) {
		return "", errors.New("worker executable is not a regular file")
	}
	if checkpoint != nil {
		if err := checkpoint(retainSourceOpened); err != nil {
			return "", err
		}
	}
	directory = filepath.Join(directory, "native")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	pattern := ".retain-"
	if runtime.GOOS == "windows" {
		pattern = ".retain-*.exe"
	}
	temp, err := os.CreateTemp(directory, pattern)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.Remove(temp.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	digest := sha256.New()
	copied, copyErr := io.Copy(io.MultiWriter(temp, digest), source)
	if copyErr != nil {
		return "", errors.Join(copyErr, temp.Close())
	}
	if checkpoint != nil {
		if err := checkpoint(retainSourceCopied); err != nil {
			return "", errors.Join(err, temp.Close())
		}
	}
	after, statErr := source.Stat()
	if statErr != nil {
		return "", errors.Join(statErr, temp.Close())
	}
	if copied != info.Size() || after.Size() != info.Size() || !after.ModTime().Equal(info.ModTime()) || !os.SameFile(info, after) {
		return "", errors.Join(errors.New("native source changed while retaining executable; pending work preserved"), temp.Close())
	}
	// Re-read the opened source inode before admission. A replaced pathname is a
	// new preference; an in-place rebuild cannot supply mixed old/new bytes.
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", errors.Join(err, temp.Close())
	}
	confirmation := sha256.New()
	confirmed, err := io.Copy(confirmation, source)
	finalInfo, finalErr := source.Stat()
	if err != nil || finalErr != nil {
		return "", errors.Join(err, finalErr, temp.Close())
	}
	if confirmed != copied || !bytes.Equal(confirmation.Sum(nil), digest.Sum(nil)) || finalInfo.Size() != info.Size() || !finalInfo.ModTime().Equal(info.ModTime()) {
		return "", errors.Join(errors.New("native source changed before retained admission"), temp.Close())
	}
	if err := temp.Sync(); err != nil {
		return "", errors.Join(err, temp.Close())
	}
	if err := temp.Chmod(0555); err != nil {
		return "", errors.Join(err, temp.Close())
	}
	if err := temp.Close(); err != nil {
		return "", err
	}
	if _, err := config.ValidateNativeExecutable(temp.Name()); err != nil {
		return "", fmt.Errorf("retained native admission failed: %w", err)
	}
	name := hex.EncodeToString(digest.Sum(nil))
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(directory, name)
	if existingInfo, err := os.Lstat(target); err == nil && !existingInfo.Mode().IsRegular() {
		return "", errors.New("retained native destination must be an owned regular file")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if existing, err := os.Open(target); err == nil {
		hash := sha256.New()
		_, readErr := io.Copy(hash, existing)
		if err := errors.Join(readErr, existing.Close()); err != nil {
			return "", err
		}
		if hex.EncodeToString(hash.Sum(nil)) != hex.EncodeToString(digest.Sum(nil)) {
			return "", fmt.Errorf("retained native artifact corrupted: %s", target)
		}
		return target, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.Rename(temp.Name(), target); err != nil {
		return "", err
	}
	return target, nil
}
