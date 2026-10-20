SELECT
    benchmark,
    iterations,
    round(ns_per_op / 1000000, 3) AS ms_per_op,
    bytes_per_op AS bytes,
    allocs_per_op AS allocs
FROM read_csv_auto(getenv('BENCH_CSV'), header = true)
ORDER BY ns_per_op;
