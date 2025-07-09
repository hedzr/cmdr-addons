package filelock

import (
	"context"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync"
	"time"
)

type Option func(f *Filelock)

// SetFlag sets the flag used to create/open the file.
func SetFlag(flag int) Option {
	return func(f *Filelock) {
		f.flag = flag
	}
}

// SetPermissions sets the OS permissions to set on the file.
func SetPermissions(perm fs.FileMode) Option {
	return func(f *Filelock) {
		f.perm = perm
	}
}

// Filelock is the struct type to handle file locking. All fields are unexported,
// with access to some of the fields provided by getter methods (Path() and Locked()).
type Filelock struct {
	path string
	m    sync.RWMutex
	fh   *os.File
	l    bool
	r    bool

	// flag is the flag used to create/open the file.
	flag int
	// perm is the OS permissions to set on the file.
	perm fs.FileMode
}

// New returns a new instance of *Filelock. The only parameter
// it takes is the path to the desired lockfile.
func New(path string, opts ...Option) *Filelock {
	// create it if it doesn't exist, and open the file read-only.
	flags := os.O_CREATE
	switch runtime.GOOS {
	case "aix", "solaris", "illumos":
		// AIX cannot preform write-lock (i.e. exclusive) on a read-only file.
		flags |= os.O_RDWR
	default:
		flags |= os.O_RDONLY
	}

	f := &Filelock{
		path: path,
		flag: flags,
		perm: fs.FileMode(0o600),
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

// NewFlock returns a new instance of *Filelock. The only parameter
// it takes is the path to the desired lockfile.
//
// Deprecated: Use New instead.
func NewFlock(path string) *Filelock {
	return New(path)
}

// Close is equivalent to calling Unlock.
//
// This will release the lock and close the underlying file descriptor.
// It will not remove the file from disk, that's up to your application.
func (f *Filelock) Close() error {
	return f.Unlock()
}

// Path returns the path as provided in NewFlock().
func (f *Filelock) Path() string {
	return f.path
}

// Locked returns the lock state (locked: true, unlocked: false).
//
// Warning: by the time you use the returned value, the state may have changed.
func (f *Filelock) Locked() bool {
	f.m.RLock()
	defer f.m.RUnlock()

	return f.l
}

// RLocked returns the read lock state (locked: true, unlocked: false).
//
// Warning: by the time you use the returned value, the state may have changed.
func (f *Filelock) RLocked() bool {
	f.m.RLock()
	defer f.m.RUnlock()

	return f.r
}

func (f *Filelock) String() string {
	return f.path
}

func (f *Filelock) WriteString(str string) (int, error) {
	return f.Write([]byte(str))
}

func (f *Filelock) Write(data []byte) (int, error) {
	return f.fh.Write(data)
}

func (f *Filelock) Read(data []byte) (int, error) {
	return f.fh.Read(data)
}

func (f *Filelock) ReadAll() (string, error) {
	data, err := io.ReadAll(f.fh)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// TryLockContext repeatedly tries to take an exclusive lock until one of the conditions is met:
// - TryLock succeeds
// - TryLock fails with error
// - Context Done channel is closed.
func (f *Filelock) TryLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, f.TryLock, retryDelay)
}

// TryRLockContext repeatedly tries to take a shared lock until one of the conditions is met:
// - TryRLock succeeds
// - TryRLock fails with error
// - Context Done channel is closed.
func (f *Filelock) TryRLockContext(ctx context.Context, retryDelay time.Duration) (bool, error) {
	return tryCtx(ctx, f.TryRLock, retryDelay)
}

func tryCtx(ctx context.Context, fn func() (bool, error), retryDelay time.Duration) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	for {
		if ok, err := fn(); ok || err != nil {
			return ok, err
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryDelay):
			// try again
		}
	}
}

func (f *Filelock) setFh(flag int) error {
	// open a new os.File instance
	fh, err := os.OpenFile(f.path, flag, f.perm)
	if err != nil {
		return err
	}

	// set the file handle on the struct
	f.fh = fh

	return nil
}

// resetFh resets file handle:
// - tries to close the file (ignore errors)
// - sets fh to nil.
func (f *Filelock) resetFh() {
	if f.fh == nil {
		return
	}

	_ = f.fh.Close()

	f.fh = nil
}

// ensure the file handle is closed if no lock is held.
func (f *Filelock) ensureFhState() {
	if f.l || f.r || f.fh == nil {
		return
	}

	f.resetFh()
}

func (f *Filelock) reset() {
	f.l = false
	f.r = false

	f.resetFh()
}
