package capture

import "golang.org/x/sys/unix"

func swapConfigEntries(first, second string) error {
	return unix.RenamexNp(first, second, unix.RENAME_SWAP)
}

func moveConfigIfVacant(from, to string) error {
	return unix.RenamexNp(from, to, unix.RENAME_EXCL)
}
