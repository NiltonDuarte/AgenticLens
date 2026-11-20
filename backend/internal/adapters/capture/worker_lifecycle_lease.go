package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"syscall"
	"time"
)

const WorkerLifecycleProtocol = 1

// WorkerLifecycleLease keeps one compatible worker elected while excluding
// older workers, which acquire worker.lock exclusively.
type WorkerLifecycleLease struct {
	compatibility *os.File
	election      *os.File
}

func AcquireLifecycleWorker(directory string) (*WorkerLifecycleLease, error) {
	election, err := tryCaptureWorkerFileLock(directory, "capture-worker-owner.lock", tryPublicationLock)
	if err != nil || election == nil {
		return nil, err
	}
	compatible, err := tryCaptureWorkerFileLock(directory, "worker.lock", trySharedPublicationLock)
	if err != nil || compatible == nil {
		return nil, errors.Join(err, election.Close())
	}
	// A normal legacy handoff may briefly outlive its exclusive lease. Give
	// that retiring group a bounded interval before replacing its progress.
	if err := checkLegacyCaptureWorkerGroup(directory, 2*time.Second); err != nil {
		return nil, errors.Join(err, compatible.Close(), election.Close())
	}
	return &WorkerLifecycleLease{compatibility: compatible, election: election}, nil
}

func (l *WorkerLifecycleLease) Close() error {
	return errors.Join(l.compatibility.Close(), l.election.Close())
}

// AcquireWorkerControlAdmission keeps older workers excluded through a direct
// lifecycle mutation, even if the currently compatible worker exits or crashes.
func AcquireWorkerControlAdmission(directory string) (func() error, error) {
	compatible, err := tryCaptureWorkerFileLock(directory, "worker.lock", trySharedPublicationLock)
	if err != nil {
		return nil, err
	}
	if compatible == nil {
		return nil, errors.New("capture worker or migration owns the legacy exclusive lease; retry after it retires")
	}
	if err := checkLegacyCaptureWorkerGroup(directory, 0); err != nil {
		return nil, errors.Join(err, compatible.Close())
	}
	return compatible.Close, nil
}

func checkLegacyCaptureWorkerGroup(directory string, wait time.Duration) error {
	file, err := openCaptureWorkerFile(directory, "worker-status.json", os.O_RDONLY)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	body, err := io.ReadAll(io.LimitReader(file, MaxPayloadBytes+1))
	if err = errors.Join(err, file.Close()); err != nil {
		return err
	}
	if len(body) > MaxPayloadBytes {
		return errors.New("recorded capture worker status exceeds its read limit")
	}
	// A corrupt binding scan is still repairable. Check only the fields
	// needed for legacy group refusal, before the status can be replaced.
	var status struct {
		PID               int `json:"pid"`
		LifecycleProtocol int `json:"lifecycle_protocol"`
	}
	if err := json.Unmarshal(body, &status); err != nil {
		// A malformed checkpoint with no recorded PID has no group to check.
		// Native legacy workers publish complete JSON atomically before launch.
		if !bytes.Contains(body, []byte(`"pid"`)) {
			return nil
		}
		return fmt.Errorf("cannot verify recorded legacy capture worker group: %w", err)
	}
	if status.PID == 0 || status.LifecycleProtocol == WorkerLifecycleProtocol {
		return nil
	}
	if status.PID <= 1 || status.PID > math.MaxInt32 {
		return errors.New("invalid recorded legacy capture worker group")
	}
	deadline := time.Now().Add(wait)
	for {
		// The supported native launcher uses Setsid. Signal zero never
		// signals a process and is used only as an additional refusal.
		err := syscall.Kill(-status.PID, 0)
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		if err != nil && !errors.Is(err, syscall.EPERM) {
			return fmt.Errorf("checking recorded legacy capture worker group %d: %w", status.PID, err)
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("recorded legacy capture worker group %d may still be active; retry after it retires", status.PID)
		}
		time.Sleep(min(20*time.Millisecond, time.Until(deadline)))
	}
}

func openCaptureWorkerFile(directory, name string, flags int) (*os.File, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	file, err := root.OpenFile(name, flags|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	closeErr := root.Close()
	if err != nil {
		return nil, errors.Join(err, closeErr)
	}
	if closeErr != nil {
		return nil, errors.Join(closeErr, file.Close())
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Join(fmt.Errorf("capture worker file %s is not a regular file", name), file.Close())
	}
	return file, nil
}

func tryCaptureWorkerFileLock(directory, name string, try func(*os.File) (bool, error)) (*os.File, error) {
	file, err := openCaptureWorkerFile(directory, name, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return nil, err
	}
	locked, err := try(file)
	if err != nil || !locked {
		return nil, errors.Join(err, file.Close())
	}
	return file, nil
}
