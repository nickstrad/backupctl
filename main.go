package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
)

func publish(path string, data []byte) error {

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

	// Release after publication finishes.
	defer syscall.Flock(
		int(lock.Fd()),
		syscall.LOCK_UN,
	)

	// Everything below runs while holding the lock.

	dir := filepath.Dir(path)

	f, err := os.CreateTemp(dir, ".snapshot-*")
	if err != nil {
		return err
	}

	tmp := f.Name()
	defer os.Remove(tmp)

	// Write snapshot.
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}

	// Synchronize file contents.
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
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

func main() {
	err := publish(
		"snapshot.txt",
		[]byte("snapshot-version=4\n"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("Snapshot published under exclusive lock")
}
