GO      ?= go
BENCH_CSV ?= bench.csv

.PHONY: all build backupctl test vet bench clean run-fd-experiment

all: build

build: backupctl

backupctl:
	$(GO) build -o backupctl ./cmd/backupctl

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

bench:
	$(GO) run ./cmd/bench -out $(BENCH_CSV)
	BENCH_CSV=$(BENCH_CSV) duckdb -no-stdin -box -f scripts/bench.sql

clean:
	rm -f backupctl $(BENCH_CSV)

run-fd-experiment:
	$(GO) run ./experiments/fd-reader.go
