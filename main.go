package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func publish(path string, data []byte) error {
	dir := filepath.Dir(path)

	// Create a unique temporary file in the same directory.
	f, err := os.CreateTemp(dir, ".snapshot-*")
	if err != nil {
		return err
	}

	tmp := f.Name()
	defer os.Remove(tmp)

	// Write the complete snapshot before publishing.
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Atomically replace the destination.
	if err := os.Rename(tmp, path); err != nil {
		return err
	}

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

func runMultiFileDescriptorFlow() {
	check(os.MkdirAll("fd-lab", 0755))

	path := "fd-lab/go-current.txt"
	tmp := "fd-lab/go-next.txt"

	// Publish the initial file.
	check(os.WriteFile(path, []byte("version=1\n"), 0644))

	// Keep a descriptor open to version 1.
	oldFD, err := os.Open(path)
	check(err)
	defer oldFD.Close()

	// Create and publish a replacement.
	check(os.WriteFile(tmp, []byte("version=2\n"), 0644))
	check(os.Rename(tmp, path))

	// Read through the original descriptor.
	oldData, err := io.ReadAll(oldFD)
	check(err)

	// Open the pathname again.
	newData, err := os.ReadFile(path)
	check(err)

	fmt.Printf("Old FD: %s", oldData)
	fmt.Printf("New open: %s", newData)
}

func runSnapshotFlow() {
	if err := publish("snapshot.txt", []byte("snapshot-version=1\n")); err != nil {
		panic(err)
	}

	fmt.Println("Snapshot published")
}

func main() {
	runMultiFileDescriptorFlow()
}
