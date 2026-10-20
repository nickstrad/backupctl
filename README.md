# backupctl

A small Go project for learning how to publish a file safely on Unix. It
writes a snapshot so that readers see either the old file or the whole new
one, never a partial write, and concurrent writers take turns.

## How publishing works

`internal/publish.Publish(path, data)` does the following:

1. Opens `<path>.lock` and takes an exclusive `flock(2)` lock, waiting if
   another writer holds it. The lock is released when the lock file is closed
   on return.
2. Creates a temp file `.<name>.tmp-*` in the same directory as `path`.
3. Writes the data, then calls `fsync` and closes the file.
4. Renames the temp file over `path`. The rename is atomic.
5. On Linux, opens the directory and calls `fsync` on it so the rename is
   durable.

If any step fails, the temp file is removed. The code uses `syscall.Flock`,
so it builds on Unix-like systems only (Linux, macOS), not Windows.

## Layout

```
cmd/backupctl/      cobra CLI: backup (publishes snapshot.txt) and restore (with tests)
cmd/bench/          benchmarks Publish and writes the result as CSV
internal/publish/   the locked atomic publish logic
scripts/bench.sql   DuckDB query that prints the benchmark CSV as a table
experiments/        standalone programs, run with `go run` (marked //go:build ignore)
```

## Requirements

- Go 1.26.4 or later (see `go.mod`)
- [DuckDB](https://duckdb.org/) CLI on your `PATH`, for `make bench` only

## Make targets

| Target                   | What it does                                                      |
| ------------------------ | ----------------------------------------------------------------- |
| `make` / `make build`    | Builds `./backupctl` in the repo root                              |
| `make test`              | `go test ./...`                                                   |
| `make vet`               | `go vet ./...`                                                    |
| `make bench`             | Runs the benchmark, writes `bench.csv`, prints it with DuckDB      |
| `make clean`             | Removes `backupctl` and `bench.csv`                                |
| `make run-fd-experiment` | Runs `experiments/fd-reader.go`                                    |

Set `BENCH_CSV` to write the benchmark somewhere else, for example
`make bench BENCH_CSV=/tmp/bench.csv`.

## Usage

Build first, then run `backupctl` from the directory that should hold
`snapshot.txt`.

```sh
make build

./backupctl backup                # publishes ./snapshot.txt
./backupctl restore restored.txt  # copies ./snapshot.txt to restored.txt
./backupctl --help
```

`restore` takes exactly one destination and refuses to copy
`snapshot.txt` onto itself. `backup` writes `snapshot-version=4` and prints
`Snapshot published under exclusive lock`; `restore` shells out to `cp` and
prints `Command completed`.

## Benchmark

```sh
make bench
```

`cmd/bench` publishes a snapshot into a temp directory in a loop using
`testing.Benchmark` and writes iterations, ns/op, bytes/op and allocs/op to
`bench.csv`. `scripts/bench.sql` then reads that file (path taken from the
`BENCH_CSV` environment variable) and prints it with ms/op.

## Tests

```sh
make test
```

The tests in `cmd/backupctl` cover help output, argument validation,
restore (including the same-path check), and failure of the child command.

## Experiments

These are single-file programs for exploring file system behavior. Both
write files into the current directory.

- `go run ./experiments/fd-reader.go` shows that an open file descriptor
  still reads the old contents after a rename replaces the path. It writes
  into `fd-lab/`.
- `go run ./experiments/stat-scan.go [path]` records a file's device, inode,
  size and mtime in `.<name>.stat.json` and reports whether they changed
  since the last scan. The default path is `snapshot.txt`.

Generated files (binaries, `bench.csv`, `snapshot.txt`, lock files, stat
caches, `fd-lab/`) are listed in `.gitignore`.
