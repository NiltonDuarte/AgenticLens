package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"agenticlens/internal/domain"
)

var errCaptureConfigBusy = errors.New("configuration admission is busy")

type captureAdmissionFile struct {
	path string
	lock *captureConfigLock
}

// CaptureAdmission serializes configured state selection and first registry
// allocation with checked path maintenance. Release it before parsing sources.
type CaptureAdmission struct {
	project string
	files   []captureAdmissionFile
}

// AcquireCaptureAdmission coordinates the shared and existing local settings.
// It never creates configuration, registry state or persistent lock artifacts.
func AcquireCaptureAdmission(ctx context.Context, project string) (*CaptureAdmission, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	shared, err := FindConfigFile(project)
	if err != nil {
		return nil, err
	}
	if shared == "" {
		return nil, ErrConfigNotFound
	}
	shared, err = filepath.Abs(shared)
	if err != nil {
		return nil, err
	}
	root := filepath.Dir(shared)
	paths := []string{shared}
	local := filepath.Join(root, ".agenticlens-config.local.yaml")
	if info, e := os.Stat(local); e == nil {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("local configuration must be a regular file: %s", local)
		}
		paths = append(paths, local)
	} else if !errors.Is(e, os.ErrNotExist) {
		return nil, e
	}
	sort.Strings(paths)
	for {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("capture configuration admission: %w", err)
		}
		admitted := &CaptureAdmission{project: root}
		var acquireErr error
		for _, path := range paths {
			lock, e := lockCaptureConfig(path)
			if e != nil {
				acquireErr = e
				break
			}
			admitted.files = append(admitted.files, captureAdmissionFile{path, lock})
			info, e := lock.Stat()
			if e != nil {
				acquireErr = e
				break
			}
			current, e := os.Stat(path)
			if e != nil {
				acquireErr = e
				break
			}
			if !current.Mode().IsRegular() || !os.SameFile(info, current) {
				acquireErr = errCaptureConfigBusy
				break
			}
		}
		if acquireErr == nil {
			if err := ctx.Err(); err != nil {
				return nil, errors.Join(err, admitted.Close())
			}
			return admitted, nil
		}
		if e := admitted.Close(); e != nil {
			return nil, errors.Join(acquireErr, e)
		}
		if !captureConfigLockBusy(acquireErr) {
			return nil, fmt.Errorf("capture configuration admission: %w", acquireErr)
		}
		timer := time.NewTimer(5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("capture configuration admission: %w", ctx.Err())
		case <-timer.C:
		}
	}
}

func (admission *CaptureAdmission) Close() error {
	var result error
	for i := len(admission.files) - 1; i >= 0; i-- {
		result = errors.Join(result, admission.files[i].lock.Close())
	}
	admission.files = nil
	return result
}

func (admission *CaptureAdmission) ResolveCapture(stateOverride string) (domain.CaptureConfig, error) {
	return ResolveCapture(admission.project, stateOverride)
}

// CaptureConfigChange contains effective state selection on either side of one
// explicit config edit, including unchanged machine-local overrides.
type CaptureConfigChange struct{ ProjectRoot, PreviousStateDir, StateDir string }

// WriteCaptureConfigUpdateChecked reruns the operational state guard while state
// admission is excluded. Fenced relocation uses WriteCaptureConfigUpdate instead.
func WriteCaptureConfigUpdateChecked(ctx context.Context, filePath string, update CaptureConfigUpdate, check func(CaptureConfigChange) error) (err error) {
	if !update.Changed {
		return nil
	}
	admission, err := AcquireCaptureAdmission(ctx, filepath.Dir(filePath))
	if err != nil {
		return err
	}
	defer func() {
		if admission != nil {
			if e := admission.Close(); e != nil {
				err = errors.Join(err, e)
			}
		}
	}()
	path, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	var target *captureConfigLock
	before, after := captureReadPolicy(), captureReadPolicy()
	shared, err := FindConfigFile(admission.project)
	if err != nil {
		return err
	}
	for _, name := range []string{shared, filepath.Join(admission.project, ".agenticlens-config.local.yaml")} {
		data, e := os.ReadFile(name)
		if errors.Is(e, os.ErrNotExist) {
			continue
		}
		if e != nil {
			return e
		}
		next := data
		info, e := os.Lstat(name)
		if e != nil {
			return e
		}
		wanted, e := os.Lstat(path)
		if e != nil {
			return e
		}
		isTarget := os.SameFile(info, wanted)
		value, present, valueErr := captureStateValue(data)
		if valueErr != nil && !isTarget {
			return valueErr
		}
		if present {
			before.StateDir = value
		}
		if isTarget {
			if !bytes.Equal(data, update.Original) {
				return errors.New("configuration changed after preview; retry without overwriting the newer file")
			}
			next = update.Data
			for _, item := range admission.files {
				lockedInfo, e := item.lock.Stat()
				if e != nil {
					return e
				}
				if os.SameFile(lockedInfo, wanted) {
					target = item.lock
					break
				}
			}
		}
		value, present, valueErr = captureStateValue(next)
		if valueErr != nil {
			return valueErr
		}
		if present {
			after.StateDir = value
		}
	}
	if err := after.Validate(); err != nil {
		return err
	}
	if target == nil {
		return errors.New("configuration update target does not belong to the admitted project")
	}
	if check != nil {
		if err = check(CaptureConfigChange{admission.project, before.StateDir, after.StateDir}); err != nil {
			return err
		}
	}
	err = writeLockedCaptureConfigUpdate(path, update, target, nil)
	closeErr := admission.Close()
	admission = nil
	return errors.Join(err, closeErr)
}
