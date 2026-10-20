package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"mini-backup-script/internal/publish"
)

func main() {
	out := flag.String("out", "bench.csv", "CSV output path")
	flag.Parse()

	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(out string) error {
	dir, err := os.MkdirTemp("", "mini-backup-bench-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "snapshot.txt")
	data := []byte("snapshot-version=4\n")

	result := testing.Benchmark(func(b *testing.B) {
		for b.Loop() {
			if err := publish.Publish(path, data); err != nil {
				b.Fatal(err)
			}
		}
	})

	f, err := os.Create(out)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	if err := w.WriteAll([][]string{
		{"benchmark", "iterations", "ns_per_op", "bytes_per_op", "allocs_per_op"},
		{
			"BenchmarkPublish",
			strconv.Itoa(result.N),
			strconv.FormatInt(result.NsPerOp(), 10),
			strconv.FormatInt(result.AllocedBytesPerOp(), 10),
			strconv.FormatInt(result.AllocsPerOp(), 10),
		},
	}); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
