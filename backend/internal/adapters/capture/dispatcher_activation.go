package capture

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ActiveDispatcher redirects the stable Windows launcher to immutable selected
// dispatcher bytes. Updating the root-owned activation record never replaces a
// running executable, including when setup was invoked through that launcher.
// Other platforms use atomic replacement and require no forwarding stage.
func ActiveDispatcher(executable string) (string, error) {
	if runtime.GOOS != "windows" {
		return "", nil
	}
	root := filepath.Dir(executable)
	data, err := readEntrypointFile(filepath.Join(root, "dispatcher-installation.json"), 1024*1024)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	var owner dispatcherInstallation
	if err := json.Unmarshal(data, &owner); err != nil {
		return "", err
	}
	if owner.Dispatcher != executable {
		return "", nil
	}
	if owner.Protocol != 1 {
		return "", errors.New("unsupported native dispatcher ownership protocol")
	}
	if owner.ActiveNative == "" {
		return "", nil
	}
	digest, err := hex.DecodeString(owner.ActiveHash)
	if err != nil || len(digest) != 32 {
		return "", errors.New("invalid active native dispatcher identity")
	}
	expected := filepath.Join(root, "native", owner.ActiveHash+".exe")
	if owner.ActiveNative != expected {
		return "", errors.New("active native dispatcher escapes owned immutable build path")
	}
	current, err := os.ReadFile(executable)
	if err != nil {
		return "", err
	}
	if hash(current) != owner.Hash {
		return "", errors.New("stable native dispatcher was modified; explicit repair required")
	}
	selected, err := os.ReadFile(expected)
	if err != nil {
		return "", err
	}
	if hash(selected) != owner.ActiveHash {
		return "", errors.New("active native dispatcher was modified; explicit repair required")
	}
	if err := validateNativeHeader(expected); err != nil {
		return "", fmt.Errorf("active native dispatcher: %w", err)
	}
	return expected, nil
}
