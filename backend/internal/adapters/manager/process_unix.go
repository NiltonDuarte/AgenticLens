//go:build !windows

package manager

import (
	"os"
	"syscall"
)

func replaceProcess(path string, args []string, directory string) (bool, error) {
	if directory != "" {
		if err := os.Chdir(directory); err != nil {
			return false, err
		}
	}
	return false, syscall.Exec(path, append([]string{path}, args...), os.Environ())
}
