package main

import (
	"fmt"
	"os"
	"path/filepath"
)

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

	if err := f.Close(); err != nil {
		return err
	}

	// Atomically replace the destination.
	return os.Rename(tmp, path)
}

func main() {
	if err := publish("snapshot.txt", []byte("snapshot-version=1\n")); err != nil {
		panic(err)
	}

	fmt.Println("Snapshot published")
}
