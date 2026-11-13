package capture

import (
	"agenticlens/internal/adapters/config"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// PublicationLock owns operational state and the actual output directory.
// The directory descriptor survives rename; no disappeared workspace is recreated.
type PublicationLock struct {
	root      *os.Root
	file      *os.File
	directory *os.File
	original  string
	state     string
	project   string
	identity  string
}

var heldPublications = struct {
	sync.Mutex
	owners map[*PublicationLock]bool
}{owners: map[*PublicationLock]bool{}}

func waitPublicationLock(ctx context.Context, file *os.File) error {
	for {
		held, err := tryPublicationLock(file)
		if err != nil {
			return err
		}
		if held {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for publication ownership: %w", ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func AcquirePublication(ctx context.Context, path string, project ...string) (*PublicationLock, error) {
	parent, err := ProjectDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	selected := ""
	if len(project) > 0 {
		selected = project[0]
	} else {
		selected = parent
	}
	selected, err = config.ProjectRoot(selected)
	if err != nil {
		return nil, err
	}
	state, err := config.ResolvePublicationState(selected)
	if err != nil {
		return nil, err
	}
	if err = MkdirPublicationStateDirectories(filepath.Join(state, "publication"), 0700); err != nil {
		return nil, fmt.Errorf("operational state home unavailable: %w", err)
	}
	directory, err := openPublicationDirectory(parent)
	if err != nil {
		return nil, fmt.Errorf("workspace moved or unavailable; resolve and retry: %w", err)
	}
	if err = waitPublicationLock(ctx, directory); err != nil {
		return nil, errors.Join(err, directory.Close())
	}
	identity, err := publicationDirectoryIdentity(directory)
	if err != nil {
		return nil, errors.Join(err, directory.Close())
	}
	digest := sha256.Sum256([]byte(selected + "\x00" + identity))
	file, err := os.OpenFile(filepath.Join(state, "publication", hex.EncodeToString(digest[:])+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, errors.Join(err, directory.Close())
	}
	if err = waitPublicationLock(ctx, file); err != nil {
		return nil, errors.Join(err, file.Close(), directory.Close())
	}
	root, err := os.OpenRoot(parent)
	if err != nil {
		return nil, errors.Join(err, file.Close(), directory.Close())
	}
	rootInfo, rootErr := root.Stat(".")
	dirInfo, dirErr := directory.Stat()
	if rootErr != nil || dirErr != nil || !os.SameFile(rootInfo, dirInfo) {
		return nil, errors.Join(fmt.Errorf("workspace replaced during directory-root admission; retry"), rootErr, dirErr, root.Close(), file.Close(), directory.Close())
	}
	owner := &PublicationLock{root: root, file: file, directory: directory, original: parent, state: state, project: selected, identity: identity}
	// A directory replaced between resolve and admission is never accepted.
	before, err := directory.Stat()
	after, e := os.Stat(parent)
	if err != nil || e != nil || !os.SameFile(before, after) {
		return nil, errors.Join(fmt.Errorf("workspace moved during admission; resolve and retry"), err, e, root.Close(), file.Close(), directory.Close())
	}
	heldPublications.Lock()
	heldPublications.owners[owner] = true
	heldPublications.Unlock()
	return owner, nil
}
func (lock *PublicationLock) Close() error {
	heldPublications.Lock()
	delete(heldPublications.owners, lock)
	heldPublications.Unlock()
	if lock.directory == nil {
		return errors.Join(unlockPublication(lock.file), lock.file.Close())
	}
	return errors.Join(unlockPublication(lock.file), lock.file.Close(), unlockPublication(lock.directory), lock.directory.Close(), lock.root.Close())
}
func (lock *PublicationLock) StateHome() string         { return lock.state }
func (lock *PublicationLock) DirectoryIdentity() string { return lock.identity }
func (lock *PublicationLock) Project() string           { return lock.project }
func (lock *PublicationLock) DirectoryPath() string     { return lock.original }
func (lock *PublicationLock) OriginalDirectory() string { return lock.original }

// PhysicalMemberPath observes a member's current name for alias and scope
// validation. Actual I/O always uses the held os.Root, never this path.
func (lock *PublicationLock) PhysicalMemberPath(path string) (string, error) {
	rel, err := publicationRelative(lock, path)
	if err != nil {
		return "", err
	}
	current, err := publicationCurrentPath(lock.directory)
	if err != nil {
		return "", err
	}
	return ProjectDirectory(filepath.Join(current, rel))
}

// PublicationOwner returns a currently held directory lease for an authorized member.
func PublicationOwner(path string) *PublicationLock {
	heldPublications.Lock()
	defer heldPublications.Unlock()
	var selected *PublicationLock
	canonical, canonicalErr := ProjectDirectory(path)
	contains := func(root, value string) bool {
		rel, err := filepath.Rel(root, value)
		return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
	}
	for owner := range heldPublications.owners {
		matches := contains(owner.original, path) || strings.HasPrefix(path, owner.DirectoryPath()+string(filepath.Separator))
		if !matches && canonicalErr == nil {
			matches = contains(owner.original, canonical)
			if !matches {
				if current, e := publicationCurrentPath(owner.directory); e == nil {
					matches = contains(current, canonical)
				}
			}
		}
		if matches && (selected == nil || len(owner.original) > len(selected.original)) {
			selected = owner
		}
	}
	return selected
}
func PublicationPath(path string) string {
	owner := PublicationOwner(path)
	if owner == nil {
		return path
	}
	rel, err := publicationRelative(owner, path)
	if err != nil {
		return path
	}
	return filepath.Join(owner.original, rel)
}

// ReadPublicationDirectory takes a read-only shared directory lease without local state.
func ReadPublicationDirectory(path string) (*ReadDirectory, error) {
	directory, err := openPublicationDirectory(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		held, e := trySharedPublicationLock(directory)
		if e != nil {
			return nil, errors.Join(e, directory.Close())
		}
		if held {
			break
		}
		select {
		case <-ctx.Done():
			return nil, errors.Join(fmt.Errorf("publication read unavailable while writer owns directory: %w", ctx.Err()), directory.Close())
		case <-time.After(20 * time.Millisecond):
		}
	}

	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, errors.Join(err, directory.Close())
	}
	rootInfo, rootErr := root.Stat(".")
	dirInfo, dirErr := directory.Stat()
	if rootErr != nil || dirErr != nil || !os.SameFile(rootInfo, dirInfo) {
		return nil, errors.Join(fmt.Errorf("workspace changed during read admission; retry"), rootErr, dirErr, root.Close(), directory.Close())
	}
	return &ReadDirectory{directory: directory, root: root, path: filepath.Dir(path)}, nil
}

// ValidateLocation verifies that the held directory still lies within selected roots.
func (lock *PublicationLock) ValidateLocation() error {
	cfg, err := config.LoadShared(lock.project)
	if err != nil {
		return err
	}
	actual, err := publicationCurrentPath(lock.directory)
	if err != nil {
		return fmt.Errorf("workspace moved or unlinked; retry resolution: %w", err)
	}
	for _, root := range cfg.Workspaces.Roots {
		selected, e := config.ResolveConfiguredPath(filepath.Dir(cfg.ConfigPath), root.Path)
		if e != nil {
			return e
		}
		selected, e = ProjectDirectory(selected)
		if e != nil {
			continue
		}
		rel, e := filepath.Rel(selected, actual)
		if e != nil || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		depth := 0
		if rel != "." {
			depth = len(strings.Split(rel, string(filepath.Separator)))
		}
		if depth <= root.Discovery.MaxDepth {
			return nil
		}
	}
	return fmt.Errorf("workspace moved outside selected root scope; resolve/retry without recreating prior location")
}
