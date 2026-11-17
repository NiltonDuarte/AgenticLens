package capture

import (
	"fmt"
	"os"
	"strings"
)

// Only current-location diagnostics/scope validation use the proc descriptor link.
// Member I/O uses os.Root and never traverses a proc path.
func publicationCurrentPath(file *os.File) (string, error) {
	path, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", file.Fd()))
	if err == nil && strings.HasSuffix(path, " (deleted)") {
		return "", fmt.Errorf("held directory has been unlinked")
	}
	return path, err
}
