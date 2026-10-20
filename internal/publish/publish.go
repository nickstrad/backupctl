package publish

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func Publish(path string, data []byte) error {

	// Open a stable lock file.
	lock, err := os.OpenFile(
		path+".lock",
		os.O_CREATE|os.O_RDWR,
		0600,
	)
	if err != nil {
		return err
	}
	defer lock.Close()

	// Acquire an exclusive lock.
	// Wait if another writer currently holds it.
	if err := syscall.Flock(
		int(lock.Fd()),
		syscall.LOCK_EX,
	); err != nil {
		return err
	}

	// The lock is released when lock is closed on return.

	dir := filepath.Dir(path)

	f, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}

	tmp := f.Name()
	defer os.Remove(tmp)

	// Write and synchronize contents, then close.
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}

	// Atomically publish.
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

	// Synchronize directory on Linux.
	if runtime.GOOS == "linux" {
		d, err := os.Open(dir)
		if err != nil {
			return err
		}
		defer d.Close()

		if err := d.Sync(); err != nil {
			return err
		}
	}

	return nil
}
