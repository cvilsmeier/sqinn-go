
SQINN-GO
==============================================================================

Sqinn-Go is a Go (Golang) library for accessing SQLite databases in pure Go.
It uses Sqinn <https://github.com/cvilsmeier/sqinn> under the hood.


Description
------------------------------------------------------------------------------

Sqinn-Go uses Sqinn for accessing SQLite database. It starts Sqinn as a child
process (`os/exec`) and communicates with Sqinn over stdin/stdout/stderr.

```go
package main

import (
	"log"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// Simple sqinn-go usage. Error handling is left out. Parameter binding is also left out.
func main() {

	// Launch sqinn. Sqinn executable must be $PATH.
	sq, _ := sqinn.New(sqinn.Options{})

	// Open a database. Database file will be created if it does not exist.
	sq.Open("./users.db")

	// Create a table.
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users.
	sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	sq.ExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

	// Query users.
	rows, _ := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.VAL_INT, sqinn.VAL_TEXT})
	for _, row := range rows {
		log.Printf("id=%d, name=%s", row.Values[0].AsInt(), row.Values[1].AsString())
		// output:
		// id=1, name=Alice
		// id=2, name=Bob
	}

	// Delete users.
	change, _ := sq.ExecOne("DELETE FROM users")
	log.Printf("deleted %d user(s)", change)
	// output:
	// deleted 2 user(s)

	// Close the database, we're done.
	sq.Close()

	// Terminate sqinn at exit. Not necessarily needed but good behavior.
	sq.Terminate()
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

	sq, err := sqinn.New(sqinn.Options{
        SqinnPath: "/path/to/sqinn",
    })

```

If do not want to use a pre-built binary, you can compile Sqinn yourself. See
<https://github.com/cvilsmeier/sqinn> for instructions.


Discussion
------------------------------------------------------------------------------

### Advantages

- No need to have gcc installed on development machine.

- Golang bult-in cross compilation works.

- Faster build speed (1s vs 3s).

- Smaller binary size (2MB vs 10MB).

- Better performance when used non-concurrently, see
  [performance.md](performance.md)


### Disadvantages

- Sqinn-Go is not a Golang `database/sql` Driver.

- Only one database connection at a time. For accessing two or more databases
  at the same time, one has to launch multiple instances of Sqinn.

- Only one prepared statement at a time. At least for now.

- Only one SQL operation at a time (no concurrency). Concurrent usage of a
  Sqinn instance has to be mutex'ed by the caller.


