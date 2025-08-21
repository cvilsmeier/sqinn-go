
![Sqinn](logo.png "Sqinn")

[![GoDoc Reference](https://pkg.go.dev/badge/github.com/cvilsmeier/sqinn-go/v2)](http://godoc.org/github.com/cvilsmeier/sqinn-go/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/cvilsmeier/sqinn-go)](https://goreportcard.com/report/github.com/cvilsmeier/sqinn-go)
[![Build Status](https://github.com/cvilsmeier/sqinn-go/actions/workflows/linux.yml/badge.svg)](https://github.com/cvilsmeier/sqinn-go/actions/workflows/linux.yml)
[![Build Status](https://github.com/cvilsmeier/sqinn-go/actions/workflows/windows.yml/badge.svg)](https://github.com/cvilsmeier/sqinn-go/actions/workflows/windows.yml)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Sqinn-Go is a Go (Golang) library for accessing SQLite databases without cgo.
It uses Sqinn <https://github.com/cvilsmeier/sqinn> under the hood.
It starts Sqinn as a child process (`os/exec`) and communicates with
Sqinn over stdin/stdout/stderr. The Sqinn child process then does the SQLite
work.

If you want SQLite but do not want cgo, Sqinn-Go can be a solution.

> [!NOTE]
> This work is sponsored by Monibot - Easy Server and Application Monitoring.
> Try out Monibot at [https://monibot.io](https://monibot.io?ref=sqinn-go).
> It's free.


Usage
------------------------------------------------------------------------------

```bash
go get -u github.com/cvilsmeier/sqinn-go/v2
```

```go
import (
	"fmt"
	"github.com/cvilsmeier/sqinn-go/v2"
)

func main() {
	// Launch sqinn, close when done.
	sq := sqinn.MustLaunch(sqinn.Options{
		Db: ":memory:", // use a transient in-memory database
	})
	defer sq.Close()
	// Create a table, cleanup when done
	sq.MustExecSql("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name TEXT)")
	defer sq.MustExecSql("DROP TABLE users")
	// Insert users
	sq.MustExecParams("INSERT INTO users (id, name) VALUES (?, ?)", 3, 2, []sqinn.Value{
		sqinn.Int32Value(1), sqinn.StringValue("Alice"),
		sqinn.Int32Value(2), sqinn.StringValue("Bob"),
		sqinn.Int32Value(3), sqinn.StringValue("Carol"),
	})
	// Query users
	rows := sq.MustQueryRows(
		"SELECT id, name FROM users WHERE id >= ? ORDER BY id",
		[]sqinn.Value{sqinn.Int32Value(0)},      // query parameters
		[]byte{sqinn.ValInt32, sqinn.ValString}, // fetch id as int, name as string
	)
	for _, values := range rows {
		fmt.Printf("user id=%d, name=%s\n", values[0].Int32, values[1].String)
	}
	// Output:
	// user id=1, name=Alice
	// user id=2, name=Bob
	// user id=3, name=Carol
}
```

For usage examples, see `examples` directory.


Building
------------------------------------------------------------------------------

The library includes a prebuilt embedded build of sqinn for linux_amd64 and
windows_amd64.

If you do not want to use a prebuilt sqinn binary, you can compile sqinn
yourself. See <https://github.com/cvilsmeier/sqinn> for instructions.
You must then specify the path to sqinn like so:

```go
sq := sqinn.MustLaunch(sqinn.Options{
	Sqinn: "/path/to/sqinn",
})
```


Pros and Cons
------------------------------------------------------------------------------

### Advantages

- No need to have gcc installed on development machines.
- Go cross compilation works.
- Faster build speed than cgo (1s vs 3s for sample program).
- Smaller binary size than cgo (2MB vs 10MB for sample program).


### Disadvantages

- No built-in connection pooling.
- Sqinn-Go is not a Golang `database/sql` Driver.
- Sqinn covers only a subset of SQLite's C APIs.


Performance
------------------------------------------------------------------------------

Performance tests show that, for many use-cases, Sqinn-Go performance is better
than cgo solutions.

See <https://github.com/cvilsmeier/go-sqlite-bench> for details.


Testing
------------------------------------------------------------------------------

Sqinn-Go comes with a large set of automated unit tests. Follow these steps to
execute all tests on linux/amd64 or windows/amd64:

Get and test Sqinn-Go

```bash
go mod init test
go get -v -u github.com/cvilsmeier/sqinn-go/v2
go test github.com/cvilsmeier/sqinn-go/v2
```


Check test coverage

```bash
go test github.com/cvilsmeier/sqinn-go/v2 -coverprofile=cover.out
go tool cover -html=cover.out
```

Test coverage is ~ 90% (as of August 2025)


Discussion
------------------------------------------------------------------------------

### Go without cgo

Sqinn-Go is Go without cgo, as it does not use cgo, nor does it depend on third-party
cgo packages. However, Sqinn-Go has a runtime dependency on Sqinn, which is a
program written in C. Sqinn has to be installed separately on each machine
where a Sqinn-Go application is executing. For this to work, Sqinn has to be
compiled for every target platform. As an alternative, pre-built Sqinn binaries
for common platforms are included in sqinn-go or can be downloaded from the
Sqinn releases page <https://github.com/cvilsmeier/sqinn/releases>.


### No database/sql driver

Database/sql is Go's default abstraction layer for SQL databases. It is widely
used and there are many third-party packages built on top of it. Sqinn-Go does
not implement the database/sql interfaces. The reason is that the sql package
provides low-level function calls to prepare statements, bind parameters, fetch
column values, and so on. Sqinn could do that, too. But, since for every
function call, Sqinn-Go has to make a inter-process communication
request/response roundtrip to a sqinn child process, this would be very slow.
Instead, Sqinn-Go provides higher-level Exec/Query interfaces that should be
used in favor of low-level fine-grained functions.


### Concurrency

Sqinn/Sqinn-Go performs well in non-concurrent as well as concurrent settings,
as shown in the performance section. However, a single Sqinn instance
is inherently single-threaded, requests are served one-after-another.

If you want true concurrency at the database level, you can spin up multiple
Sqinn instances. You may even implement a connection pool. But be aware that
when accessing a SQLite database concurrently, the dreaded SQLITE_BUSY error
might occur. The PRAGMA busy_timeout might help to avoid SQLITE_BUSY errors.



Changelog
------------------------------------------------------------------------------

### v2.0.2

- better prebuilts (gzip and build constraints)


### v2.0.1

- ValXy const is now byte (was untyped)


### v2.0.0

- Major Version 2 (streaming protocol, less memory, faster)
- Include prebuilt sqinn v2.0.0 (SQLite v3.50.4 (2025-07-30))


### v1.2.0 (2023-10-05)

- Added marshalling benchmark
- Removed 'pure Go' claim
- Removed travis build
- Added github workflow with sqinn v1.1.27
- Updated min. go version 1.19
- Updated samples


### v1.1.2 (2021-05-27)

- Fixed negative int32 marshalling


### v1.1.1 (2021-03-27)

- Added more docs for Values
- Added example for handling NULL values
- Added example for sqlite specialties


### v1.1.0 (2020-06-14)

- Use IEEE 745 encoding for float64 values, needs sqinn v1.1.0 or higher.


### v1.0.0 (2020-06-10)

- First version.
