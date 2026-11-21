package live

import "context"

// readLock serializes access to a shared reader without leaving a cancelled
// caller waiting for the current operation to finish.
type readLock chan struct{}

func (lock readLock) Lock() {
	lock <- struct{}{}
}

func (lock readLock) LockContext(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case lock <- struct{}{}:
		if err := ctx.Err(); err != nil {
			lock.Unlock()
			return err
		}
		return nil
	}
}

func (lock readLock) Unlock() {
	<-lock
}
