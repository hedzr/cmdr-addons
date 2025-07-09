//go:build (!windows && !darwin && !unix) || plan9

package filelock

import (
	"errors"
	"io/fs"
)

func (f *Filelock) Lock() error {
	return &fs.PathError{
		Op:   "Lock",
		Path: f.Path(),
		Err:  errors.ErrUnsupported,
	}
}

func (f *Filelock) RLock() error {
	return &fs.PathError{
		Op:   "RLock",
		Path: f.Path(),
		Err:  errors.ErrUnsupported,
	}
}

func (f *Filelock) Unlock() error {
	return &fs.PathError{
		Op:   "Unlock",
		Path: f.Path(),
		Err:  errors.ErrUnsupported,
	}
}

func (f *Filelock) TryLock() (bool, error) {
	return false, f.Lock()
}

func (f *Filelock) TryRLock() (bool, error) {
	return false, f.RLock()
}
