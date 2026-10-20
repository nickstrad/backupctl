GO      ?= go
BENCH_CSV ?= bench.csv

.PHONY: all build mini-backup backupctl test bench clean run-fd-experiment

all: build

build: mini-backup backupctl

mini-backup:
	$(GO) build -o mini-backup .

backupctl:
	$(GO) build -o backupctl ./cmd/backupctl

test:
	$(GO) test ./...

bench:
	$(GO) run ./cmd/bench -out $(BENCH_CSV)
	BENCH_CSV=$(BENCH_CSV) duckdb -no-stdin -box -f scripts/bench.sql

clean:
	rm -f mini-backup backupctl $(BENCH_CSV)

run-fd-experiment:
	$(GO) run ./experiments/fd-reader.go
