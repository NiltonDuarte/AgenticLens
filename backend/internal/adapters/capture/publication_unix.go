//go:build linux || darwin

package capture

import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

func tryPublicationLock(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return false, nil
	}
	return err == nil, err
}
func unlockPublication(file *os.File) error { return unix.Flock(int(file.Fd()), unix.LOCK_UN) }

func openPublicationDirectory(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}
func publicationDirectoryIdentity(file *os.File) (string, error) {
	var st unix.Stat_t
	if err := unix.Fstat(int(file.Fd()), &st); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d-%d", st.Dev, st.Ino), nil
}
func trySharedPublicationLock(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_SH|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return false, nil
	}
	return err == nil, err
}
