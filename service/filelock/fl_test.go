package filelock

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlock_fh_onError(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "go-flock-")
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	err = os.Remove(tmpFile.Name())
	require.NoError(t, err)

	lock := New(tmpFile.Name())

	locked, err := lock.TryLock()
	require.NoError(t, err)
	require.True(t, locked)

	newLock := New(tmpFile.Name())

	locked, err = newLock.TryLock()
	require.NoError(t, err)
	require.False(t, locked)

	assert.Nil(t, newLock.fh, "file handle should have been released and be nil")

	err = lock.Unlock()
	require.NoError(t, err)
}

func TestFlock_fh_onError_dir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on Windows")
	}

	tmpDir := t.TempDir()

	lock := New(tmpDir, SetFlag(os.O_RDONLY))

	locked, err := lock.TryLock()
	require.NoError(t, err)
	require.True(t, locked)

	newLock := New(tmpDir, SetFlag(os.O_RDONLY))

	locked, err = newLock.TryLock()
	require.NoError(t, err)
	require.False(t, locked)

	assert.Nil(t, newLock.fh, "file handle should have been released and be nil")

	err = lock.Unlock()
	require.NoError(t, err)
}
