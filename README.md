
![Sqinn](logo.png "Sqinn")

[![GoDoc Reference](https://godoc.org/github.com/cvilsmeier/sqinn-go/sqinn?status.svg)](http://godoc.org/github.com/cvilsmeier/sqinn-go/sqinn)
[![Go Report Card](https://goreportcard.com/badge/github.com/cvilsmeier/sqinn-go)](https://goreportcard.com/report/github.com/cvilsmeier/sqinn-go)
[![Build Status](https://github.com/cvilsmeier/sqinn-go/actions/workflows/linux.yml/badge.svg)](https://github.com/cvilsmeier/sqinn-go/actions/workflows/linux.yml)
[![Build Status](https://github.com/cvilsmeier/sqinn-go/actions/workflows/windows.yml/badge.svg)](https://github.com/cvilsmeier/sqinn-go/actions/workflows/windows.yml)
[![License: Unlicense](https://img.shields.io/badge/license-Unlicense-blue.svg)](http://unlicense.org/)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

Sqinn-Go is a Go (Golang) library for accessing SQLite databases without cgo.
It uses Sqinn2 <https://github.com/cvilsmeier/sqinn2> under the hood.
It starts Sqinn2 as a child process (`os/exec`) and communicates with
Sqinn2 over stdin/stdout/stderr. The Sqinn2 child process then does the SQLite
work.

If you want SQLite but do not want cgo, Sqinn-Go can be a solution.

> [!NOTE]
> This work is sponsored by Monibot - Easy Server and Application Monitoring.
> Try out Monibot at [https://monibot.io](https://monibot.io?ref=sqinn-go).
> It's free.


Usage
------------------------------------------------------------------------------

```
$ go get -u github.com/cvilsmeier/sqinn-go/v2
```

```go
import (
	"fmt"
	"github.com/cvilsmeier/sqinn-go/v2"
)

func main() {
	// Launch sqinn.
	sq := sqinn.MustLaunch(sqinn.Options{
		Db: ":memory:", // use a transient in-memory database
	})
	defer sq.Close()
	// Create a table, cleanup when done
	sq.MustExecSql("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")
	defer sq.MustExecSql("DROP TABLE users")
	// Insert users
	sq.MustExec("INSERT INTO users (id, name) VALUES (?, ?)", [][]any{
		{1, "Alice"},
		{2, "Bob"},
		{3, "Carol"},
	})
	// Query users
	rows := sq.MustQuery(
		"SELECT id, name FROM users WHERE id >= ? ORDER BY id",
		[]any{0},                                // query parameters
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

The library uses a pre-built embedded build of sqinn2 for Linux/amd64 and
Windows/amd64.

If you do not want to use a pre-built sqinn2 binary, you can compile sqinn2
yourself. See <https://github.com/cvilsmeier/sqinn2> for instructions.
You must then specify the path to sqinn2 like so:

```go
    sq := sqinn.MustLaunch(sqinn.Options{
        Sqinn2: "/path/to/sqinn2",
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

See <https://github.com/cvilsmeier/sqinn-go-bench> for details.


Testing
------------------------------------------------------------------------------

Sqinn-Go comes with a large set of automated unit tests. Follow these steps to
execute all tests on linux/amd64 or windows/amd64:

Get and test Sqinn-Go

	$ go get -v -u github.com/cvilsmeier/sqinn-go/v2
	$ go test github.com/cvilsmeier/sqinn-go/v2

Check test coverage

	$ go test github.com/cvilsmeier/sqinn-go/v2 -coverprofile=./cover.out
	$ go tool cover -html=./cover.out

Test coverage is ~90% (as of August 2025)


Discussion
------------------------------------------------------------------------------

### Go without cgo

Sqinn-Go is Go without cgo, as it does not use cgo, nor does it depend on third-party
cgo packages. However, Sqinn-Go has a runtime dependency on Sqinn, which is a
program written in C. Sqinn has to be installed separately on each machine
where a Sqinn-Go application is executing. For this to work, Sqinn has to be
compiled for every target platform. As an alternative, pre-built Sqinn binaries
for common platforms can be downloaded from the Sqinn releases page
<https://github.com/cvilsmeier/sqinn/releases>.


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
as shown in the Performance section. However, a single Sqinn instance
should only be called from one goroutine. Exceptions are the Exec and Query
methods, these are mutex'ed and goroutine safe. But, since Sqinn is inherently
single-threaded, Exec and Query requests are served one-after-another.

If you want true concurrency at the database level, you can spin up multiple
Sqinn instances. You may even implement a connection pool. But be aware that
when accessing a SQLite database concurrently, the dreaded SQLITE_BUSY error
might occur. The PRAGMA busy_timeout might help to avoid SQLITE_BUSY errors.

We recommend the following: Have one Sqinn instance. You may call Exec/Query on
that single Sqinn instance from as many goroutines as you want. For
long-running tasks (VACUUM, BACKUP, etc), spin up a second Sqinn instance on
demand, and terminate it once the long-running work is done. Use PRAGMA
busy_timeout to avoid SQLITE_BUSY.


### Only one active statement at a time

A Sqinn instance allows only one active statement at a time. A statement is
*active* from the time it is prepared until it is finalized.  Before preparing
a new statement, you have to finalize the current statement first, otherwise
Sqinn will respond with an error.

This is why we recommend using Exec/Query: These methods do a complete
prepare-finalize cycle and the caller can be sure that, once Exec/Query
returns, no active statements are hanging around.


Changelog
------------------------------------------------------------------------------

### v2.0.0

- Major Version 2 (less memory, faster)
- Uses https://github.com/cvilsmeier/sqinn2


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


License
------------------------------------------------------------------------------

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or distribute this
software, either in source code form or as a compiled binary, for any purpose,
commercial or non-commercial, and by any means.

In jurisdictions that recognize copyright laws, the author or authors of this
software dedicate any and all copyright interest in the software to the public
domain. We make this dedication for the benefit of the public at large and to
the detriment of our heirs and successors. We intend this dedication to be an
overt act of relinquishment in perpetuity of all present and future rights to
this software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <https://unlicense.org>

