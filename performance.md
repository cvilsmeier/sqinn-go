
PERFORMANCE
==============================================================================

Performance tests show that Sqinn-Go performs a little better than cgo based
alternatives.

For performance comparison I used <https://github.com/mattn/go-sqlite3>.  The
test code can be found in the example subdirectory. Please note that the mattn
go-sqlite3 test code is commented out, to avoid having a dependency on
go-sqlite3. If you want to execute the performance tests you can remove the
comments and then compile the program.

Please note that the test programs use one goroutine to access the database.

Here is the test setup:

- OS: Windows 10 Home x64 Version 1909 Build 18363
- CPU: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz, 2592 MHz, 4 Cores
- RAM: 16GB
- Disk: 256GB SSD

## Benchmark 1

Inserting and querying 1 million rows. The schema is only one table:

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
	| go-sqlite3       | 2.8 s  | 2.2 s  | 
	| sqinn-go         | 2.4 s  | 2.0 s  | 
	+------------------+--------+--------+


## Benchmark 2

A more complex table schema with foreign key constraints and many indices.  See
example/test-sqinn for details.

The results are (lower is better):

	+------------------+--------+--------+
	|                  | insert | query  |
	+------------------+--------+--------+
	| go-sqlite3       | 1.1 s  | 0.8 s  | 
	| sqinn-go         | 0.9 s  | 0.7 s  | 
	+------------------+--------+--------+


## Summary

In non-concurrent environments, Sqinn-Go is on par with cgo based alternatives.
However, when multiple goroutines access one database concurrently, cgo
libraries may perform better.

