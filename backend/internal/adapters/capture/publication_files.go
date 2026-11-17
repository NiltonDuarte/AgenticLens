package capture

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PublicationFile operations use os.Root's descriptor-relative containment.
// Logical paths remain stable across renames and are never used as write authority.
func publicationRelative(owner *PublicationLock, path string) (string, error) {
	rel, err := filepath.Rel(owner.original, path)
	if err != nil || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		actual, e := ProjectDirectory(path)
		if e != nil {
			return "", e
		}
		current, e := publicationCurrentPath(owner.directory)
		if e != nil {
			return "", e
		}
		rel, err = filepath.Rel(current, actual)
	}
	if err != nil || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("publication member escapes held directory")
	}
	return rel, nil
}
func OpenPublicationFile(path string, flags int, mode os.FileMode) (*os.File, error) {
	owner := PublicationOwner(path)
	if owner == nil {
		return os.OpenFile(path, flags, mode)
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return nil, err
	}
	return owner.root.OpenFile(rel, flags, mode)
}
func StatPublicationFile(path string) (os.FileInfo, error) {
	owner := PublicationOwner(path)
	if owner == nil {
		return os.Stat(path)
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return nil, err
	}
	return owner.root.Stat(rel)
}
func LstatPublicationFile(path string) (os.FileInfo, error) {
	owner := PublicationOwner(path)
	if owner == nil {
		return os.Lstat(path)
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return nil, err
	}
	return owner.root.Lstat(rel)
}
func ReadPublicationFile(path string) ([]byte, error) {
	file, err := OpenPublicationFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(file, 128<<20+1))
	err = errors.Join(readErr, file.Close())
	if err != nil {
		return nil, err
	}
	if len(data) > 128<<20 {
		return nil, fmt.Errorf("publication member exceeds bounded read")
	}
	return data, nil
}
func WritePublicationFile(path string, data []byte, mode os.FileMode) error {
	file, err := OpenPublicationFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return errors.Join(err, file.Close())
}
func RemovePublicationFile(path string) error {
	owner := PublicationOwner(path)
	if owner == nil {
		return os.Remove(path)
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return err
	}
	return owner.root.Remove(rel)
}
func RenamePublicationFile(oldpath, newpath string) error {
	owner := PublicationOwner(oldpath)
	other := PublicationOwner(newpath)
	if owner == nil && other == nil {
		return os.Rename(oldpath, newpath)
	}
	if owner == nil || owner != other {
		return fmt.Errorf("publication replacement must stay within one held directory")
	}
	oldrel, err := publicationRelative(owner, oldpath)
	if err != nil {
		return err
	}
	newrel, err := publicationRelative(owner, newpath)
	if err != nil {
		return err
	}
	return owner.root.Rename(oldrel, newrel)
}
func CreatePublicationTemp(directory, prefix string) (*os.File, error) {
	owner := PublicationOwner(directory)
	if owner == nil {
		return os.CreateTemp(directory, prefix)
	}
	for attempt := 0; attempt < 20; attempt++ {
		token := make([]byte, 12)
		if _, err := rand.Read(token); err != nil {
			return nil, err
		}
		name := filepath.Join(directory, prefix+hex.EncodeToString(token))
		file, err := OpenPublicationFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		return file, err
	}
	return nil, fmt.Errorf("unable to allocate unique publication stage")
}

// ReadDirectory holds a shared read lease and a descriptor-backed containment root.
type ReadDirectory struct {
	directory *os.File
	root      *os.Root
	path      string
}

func (d *ReadDirectory) Close() error                           { return errors.Join(d.root.Close(), d.directory.Close()) }
func (d *ReadDirectory) Open(name string) (*os.File, error)     { return d.root.Open(name) }
func (d *ReadDirectory) Stat(name string) (os.FileInfo, error)  { return d.root.Stat(name) }
func (d *ReadDirectory) Lstat(name string) (os.FileInfo, error) { return d.root.Lstat(name) }
func (d *ReadDirectory) Path() string                           { return d.path }
func (d *ReadDirectory) ReadFile(name string) ([]byte, error) {
	f, err := d.Open(name)
	if err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(io.LimitReader(f, 128<<20+1))
	err = errors.Join(readErr, f.Close())
	if err != nil {
		return nil, err
	}
	if len(body) > 128<<20 {
		return nil, fmt.Errorf("workspace member exceeds read bound")
	}
	return body, nil
}

func MkdirPublicationDirectories(path string, mode os.FileMode) error {
	owner := PublicationOwner(path)
	if owner == nil {
		return os.MkdirAll(path, mode)
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return err
	}
	return owner.root.MkdirAll(rel, mode)
}

// MkdirPublicationStateDirectories persists private recovery ancestry before any
// public pending indicator can authorize changing accepted artifact bytes.
func MkdirPublicationStateDirectories(path string, mode os.FileMode) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving publication state directory: %w", err)
	}
	missing := []string{}
	ancestor := path
	for {
		info, statErr := os.Stat(ancestor)
		if statErr == nil {
			if !info.IsDir() {
				return fmt.Errorf("publication state path is not a directory: %s", ancestor)
			}
			break
		}
		if !errors.Is(statErr, os.ErrNotExist) {
			return fmt.Errorf("inspecting publication state directory %s: %w", ancestor, statErr)
		}
		missing = append(missing, ancestor)
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return fmt.Errorf("publication state has no available ancestor: %w", statErr)
		}
		ancestor = parent
	}
	// Also persist the existing ancestor's entry. A prior failed initialization
	// may have created it but failed to flush its parent before returning.
	if err := syncPublicationStateDirectory(ancestor); err != nil {
		return err
	}
	if err := syncPublicationStateDirectory(filepath.Dir(ancestor)); err != nil {
		return err
	}
	for i := len(missing) - 1; i >= 0; i-- {
		directory := missing[i]
		if err := os.Mkdir(directory, mode); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("creating publication state directory %s: %w", directory, err)
		}
		if err := syncPublicationStateDirectory(directory); err != nil {
			return err
		}
		if err := syncPublicationStateDirectory(filepath.Dir(directory)); err != nil {
			return err
		}
	}
	return nil
}

func syncPublicationStateDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening publication state directory %s: %w", path, err)
	}
	info, err := directory.Stat()
	if err == nil && !info.IsDir() {
		err = fmt.Errorf("publication state path is not a directory: %s", path)
	}
	if err == nil {
		err = directory.Sync()
	}
	if err = errors.Join(err, directory.Close()); err != nil {
		return fmt.Errorf("syncing publication state directory %s: %w", path, err)
	}
	return nil
}
