package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

type Stamp struct {
	Device     uint64 `json:"device"`
	Inode      uint64 `json:"inode"`
	Size       int64  `json:"size"`
	ModifiedNs int64  `json:"modified_ns"`
}

func main() {
	path := "snapshot.txt"

	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	// Ask the OS for filesystem metadata.
	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	stat := info.Sys().(*syscall.Stat_t)

	current := Stamp{
		Device:     uint64(stat.Dev),
		Inode:      uint64(stat.Ino),
		Size:       info.Size(),
		ModifiedNs: info.ModTime().UnixNano(),
	}

	cache := filepath.Join(
		filepath.Dir(path),
		"."+filepath.Base(path)+".stat.json",
	)

	previousBytes, err := os.ReadFile(cache)

	if errors.Is(err, os.ErrNotExist) {
		fmt.Println("CHANGED: no previous scan")
	} else if err != nil {
		panic(err)
	} else {
		var previous Stamp

		if err := json.Unmarshal(previousBytes, &previous); err != nil {
			panic(err)
		}

		if current == previous {
			fmt.Println("UNCHANGED: metadata matches")
		} else {
			fmt.Println("CHANGED: metadata differs")
		}
	}

	// Remember the current observation for the next scan.
	nextBytes, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile(cache, nextBytes, 0600); err != nil {
		panic(err)
	}

	fmt.Printf(
		"device=%d inode=%d size=%d mtime_ns=%d\n",
		current.Device,
		current.Inode,
		current.Size,
		current.ModifiedNs,
	)
}
