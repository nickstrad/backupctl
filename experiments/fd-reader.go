//go:build ignore

package main

import (
	"fmt"
	"io"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
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

func main() {
	runMultiFileDescriptorFlow()
}
