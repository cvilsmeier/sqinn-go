
![Sqinn](logo-200.png "Sqinn")

[![GoDoc Reference](https://godoc.org/github.com/cvilsmeier/sqinn-go?status.svg)](http://godoc.org/github.com/cvilsmeier/sqinn-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cvilsmeier/sqinn-go)](https://goreportcard.com/report/github.com/cvilsmeier/sqinn-go)
[![Build Status](https://api.travis-ci.com/cvilsmeier/sqinn-go.svg?branch=master)](https://travis-ci.org/cvilsmeier/sqinn-go)

Sqinn-Go is a Go (Golang) library for accessing SQLite databases in pure Go.
It uses Sqinn <https://github.com/cvilsmeier/sqinn> under the hood.

If you want SQLite but do not want cgo, sqinn can be a solution.

Description
------------------------------------------------------------------------------

Sqinn-Go uses Sqinn for accessing SQLite database. It starts Sqinn as a child
process (`os/exec`) and communicates with Sqinn over stdin/stdout/stderr.

```go
import "github.com/cvilsmeier/sqinn-go/sqinn"

// Simple sqinn-go usage. Error handling is left out for brevity.
func main() {
	
	// Launch sqinn. Terminate at program exit
	sq, _ := sqinn.Launch(sqinn.Options{})
	defer sq.Terminate()

	// Open database. Close when we're done.
	sq.Open("./users.db")
	defer sq.Close()

	// Create a table.
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users.
	sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	sq.ExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

	// Query users.
	rows, _ := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
	for _, row := range rows {
		fmt.Printf("id=%d, name=%s\n", row.Values[0].AsInt(), row.Values[1].AsString())
	}

	// Output:
	// id=1, name=Alice
	// id=2, name=Bob
}
```


Testing
------------------------------------------------------------------------------

Sqinn-Go comes with a large set of automated unit tests. Follow these steps to
execute all tests on linux_amd64:

Download and Install Sqinn

	$ cd $HOME
	$ curl -sL https://github.com/cvilsmeier/sqinn/releases/download/v1.0.0/sqinn-dist-1.0.0.tar.gz | tar xz
	$ export SQINN_PATH=$HOME/sqinn-dist-1.0.0/linux_amd64/sqinn

Get and test Sqinn-Go

	$ go get -v -u github.com/cvilsmeier/sqinn-go/...
	$ go test -v github.com/cvilsmeier/sqinn-go/...

Check test coverage

	$ go test -v github.com/cvilsmeier/sqinn-go/... -coverprofile=./cover.out
	$ go tool cover -func=./cover.out	
	$ go tool cover -html=./cover.out

Test coverage is ~85% (as of 2020-06-10)
	



Usage
------------------------------------------------------------------------------

Sqinn must be installed on your system. The easiest way is to download a
pre-built executable from <https://github.com/cvilsmeier/sqinn/releases> and
put it somewhere on your `$PATH`, or `%PATH%` on Windows.

If you want to store the Sqinn binary in a non-path folder, you can do that.
But then you must specify it when opening a Sqinn connection:

```go

    // use explicit path...
    sq, _ := sqinn.New(sqinn.Options{
        SqinnPath: "/path/to/sqinn",
    })

    // ...or take from environment
    sq, _ := sqinn.New(sqinn.Options{
        SqinnPath: os.Getenv("SQINN_PATH"),
    })

```

If do not want to use a pre-built binary, you can compile Sqinn yourself. See
<https://github.com/cvilsmeier/sqinn> for instructions.


Pros and Cons
------------------------------------------------------------------------------

### Advantages

- No need to have gcc installed on development machine.

- Golang cross compilation works.

- Faster build speed (1s vs 3s for sample program).

- Smaller binary size (2MB vs 10MB for sample program).

- Better performance than cgo solutions, see [performance.md](performance.md)


### Disadvantages

- No out-of-the-box connection pooling.

- Sqinn-Go is not a Golang `database/sql` Driver.


Discussion
------------------------------------------------------------------------------

### Pure Go, almost

Sqinn-Go is pure Go, as it does not use cgo, nor does it depend on third-party
cgo packages. However, Sqinn-Go has a runtime dependency on Sqinn, which is a
program written in C. Sqinn has to be installed separately on each machine
where a Sqinn-Go application is executing. For this to work, Sqinn has to be
compiled for every target platform. As an alternative, pre-built Sqinn binaries
for the most common platforms can be downloaded from the Sqinn releases page
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
as shown in [performance.md](performance.md). However, a single Sqinn instance
should only be called from one goroutine. Exceptions are the Exec and Query
methods, these are mutex'ed and goroutine safe. But, since Sqinn is inherently
single-threaded, Exec and Query requests are served one-after-another.

If you want true concurrency at the database level, you can spin up multiple
Sqinn instances. You may even implement a connection pool. But be aware that
when accessing a SQLite database concurrently, the dreaded SQLITE_BUSY error
might occur.

We recommend the following: Have one Sqinn instance. You may call Exec/Query on
that single Sqinn instance from as many goroutines as you want.  For
long-running tasks (VACUUM, BACKUP, etc), spin up a second Sqinn instance on
demand, and terminate it once the long-running work is done. Use PRAGMA
busy_timeout to avoid SQLITE_BUSY.


### Only one active statement at a time

A Sqinn instance allows only one active statement at a time. A statement is
/active/ from the time it is prepared until it is finalized.  Before preparing
a new statement, you have to finalize the current statement first, otherwise
Sqinn will respond with an error.

This is why we recommend using Exec/Query: These methods do a complete
prepare-finalize cycle and the caller can be sure that, once Exec/Query
returns, no active statements are hanging around.



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

