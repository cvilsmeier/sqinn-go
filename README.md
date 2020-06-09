
SQINN-GO
==============================================================================

Sqinn-Go is a Go (Golang) library for accessing SQLite databases in pure Go.
It uses Sqinn <https://github.com/cvilsmeier/sqinn> under the hood.

If you want SQLite but do not want cgo, sqinn can be a solution.


    !!!

    Not production-ready. 

    Preliminary version, everything may change.

    !!!

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


Discussion
------------------------------------------------------------------------------

### Advantages

- No need to have gcc installed on development machine.

- Golang cross compilation works.

- Faster build speed (1s vs 3s for sample program).

- Smaller binary size (2MB vs 10MB for sample program).

- Better performance than cgo solutions, see
  [performance.md](performance.md)


### Disadvantages

- No out-of-the-box connection pooling.

- Sqinn-Go is not a Golang `database/sql` Driver.


License
------------------------------------------------------------------------------

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

In jurisdictions that recognize copyright laws, the author or authors
of this software dedicate any and all copyright interest in the
software to the public domain. We make this dedication for the benefit
of the public at large and to the detriment of our heirs and
successors. We intend this dedication to be an overt act of
relinquishment in perpetuity of all present and future rights to this
software under copyright law.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

For more information, please refer to <https://unlicense.org>

