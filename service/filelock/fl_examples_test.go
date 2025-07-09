package filelock_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hedzr/cmdr-addons/v2/service/v2/filelock"
)

func ExampleFlock_Locked() {
	f := filelock.New(filepath.Join(os.TempDir(), "go-lock.lock"))

	_, err := f.TryLock()
	if err != nil {
		// handle locking error
		panic(err)
	}

	fmt.Printf("locked: %v\n", f.Locked())

	err = f.Unlock()
	if err != nil {
		// handle locking error
		panic(err)
	}

	fmt.Printf("locked: %v\n", f.Locked())

	// Output: locked: true
	// locked: false
}

func ExampleFlock_TryLock() {
	// should probably put these in /var/lock
	f := filelock.New(filepath.Join(os.TempDir(), "go-lock.lock"))

	locked, err := f.TryLock()
	if err != nil {
		// handle locking error
		panic(err)
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", f.Path(), f.Locked())

		if err := f.Unlock(); err != nil {
			// handle unlock error
			panic(err)
		}
	}

	fmt.Printf("path: %s; locked: %v\n", f.Path(), f.Locked())
}

func ExampleFlock_TryLockContext() {
	// should probably put these in /var/lock
	f := filelock.New(filepath.Join(os.TempDir(), "go-lock.lock"))

	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	locked, err := f.TryLockContext(lockCtx, 678*time.Millisecond)
	if err != nil {
		// handle locking error
		panic(err)
	}

	if locked {
		fmt.Printf("path: %s; locked: %v\n", f.Path(), f.Locked())

		if err := f.Unlock(); err != nil {
			// handle unlock error
			panic(err)
		}
	}

	fmt.Printf("path: %s; locked: %v\n", f.Path(), f.Locked())
}
