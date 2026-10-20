package main

import (
	"fmt"
	"mini-backup-script/internal/publish"
	"os"
)

func main() {
	err := publish.Publish(
		"snapshot.txt",
		[]byte("snapshot-version=4\n"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("Snapshot published under exclusive lock")
}
