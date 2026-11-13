package capture

import (
	"bytes"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"runtime"
	"unsafe"
)

// F_GETPATH observes the descriptor's current name only for scope revalidation.
// Public member operations use os.Root's descriptor-backed containment.
func publicationCurrentPath(file *os.File) (string, error) {
	buffer := make([]byte, unix.PathMax)
	var pinned runtime.Pinner
	pinned.Pin(&buffer[0])
	defer pinned.Unpin()
	_, err := unix.FcntlInt(file.Fd(), unix.F_GETPATH, int(uintptr(unsafe.Pointer(&buffer[0]))))
	runtime.KeepAlive(file)
	runtime.KeepAlive(buffer)
	if err != nil {
		return "", err
	}
	end := bytes.IndexByte(buffer, 0)
	if end <= 0 {
		return "", fmt.Errorf("directory descriptor has no current path")
	}
	return string(buffer[:end]), nil
}
