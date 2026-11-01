package capture

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PublicationLock serializes the entire baseline-read/extract/publish operation
// for a ledger, including generations that do not use a capture registry.
// Kernel ownership is released on process exit/crash. The inode is retained;
// unlinking it would allow two concurrent owners after a waiting process opens it.
type PublicationLock struct{ file *os.File }

func AcquirePublication(ctx context.Context, path string) (*PublicationLock, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	parent := filepath.Dir(absolute)
	if err = os.MkdirAll(parent, 0755); err != nil {
		return nil, err
	}
	resolved, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(resolved, filepath.Base(absolute))
	if leaf, e := filepath.EvalSymlinks(target); e == nil {
		target = leaf
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	lockDir := filepath.Join(cache, "agenticlens", "publication")
	if err = os.MkdirAll(lockDir, 0700); err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(target))
	lockPath := filepath.Join(lockDir, hex.EncodeToString(digest[:])+".agenticlens-lock")
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	for {
		held, e := tryPublicationLock(file)
		if e != nil {
			return nil, errors.Join(e, file.Close())
		}
		if held {
			return &PublicationLock{file}, nil
		}
		select {
		case <-ctx.Done():
			return nil, errors.Join(fmt.Errorf("waiting for ledger publication ownership: %w", ctx.Err()), file.Close())
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func (lock *PublicationLock) Close() error {
	return errors.Join(unlockPublication(lock.file), lock.file.Close())
}
