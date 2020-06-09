
PERFORMANCE
==============================================================================

Performance tests show that Sqinn-Go performs slightly better than cgo based
alternatives.

For benchmarks I used <https://github.com/mattn/go-sqlite3>. It's been around a
long time, well tested, and widely used.

The benchmarking code can be found in the examples subdirectory. Please note
that the mattn go-sqlite3 test code is commented out, to avoid having a
compile-time dependency on go-sqlite3. If you want to compile and execute the
benchmarks you have to re-enable the commented lines first.

The test setup is as follows:

- OS: Windows 10 Home x64 Version 1909 Build 18363
- CPU: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz, 2592 MHz, 4 Cores
- RAM: 16GB
- Disk: 256GB SSD

## Benchmark 1

Inserting and querying 1 million rows in one goroutine. The schema is only one
table:

	CREATE TABLE users (
		id INTEGER PRIMARY KEY NOT NULL,
		name VARCHAR,
		age INTEGER,
		rating REAL
	);

The results are (lower is better):

	+------------------+--------+--------+
	|                  | insert | query  |
	+------------------+--------+--------+
	| go-sqlite3       | 2.8 s  | 2.3 s  | 
	| sqinn-go         | 2.5 s  | 2.1 s  | 
	+------------------+--------+--------+


## Benchmark 2

A more complex table schema with foreign key constraints and many indices.
Inserting and querying 200000 rows in one goroutine.

The results are (lower is better):

	+------------------+--------+--------+
	|                  | insert | query  |
	+------------------+--------+--------+
	| go-sqlite3       | 2.1 s  | 1.7 s  | 
	| sqinn-go         | 1.8 s  | 1.3 s  | 
	+------------------+--------+--------+


## Benchmark 3

Querying a table with 1 million rows concurrently. Spin up N goroutines, where
each goroutine queries all 1000000 rows.

The results are (lower is better), N is the number of goroutines:

	+------------------+--------+--------+--------+
	|                  | N=2    | N=4    | N=8    |
	+------------------+--------+--------+--------+
	| go-sqlite3       | 1.4 s  | 1.6 s  | 2.3 s  |
	| sqinn-go         | 0.9 s  | 1.1 s  | 2.1 s  |
	+------------------+--------+--------+--------+


## Summary

In every benchmark I executed, Sqinn-Go is slightly faster than the cgo
package. But: Every application is different, and I recommend that you perform
benchmarks based on the typical workload of your application. As always, it
depends.
