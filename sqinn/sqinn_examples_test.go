package sqinn_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// NOTE: For running the examples you must have sqinn installed
// and $SQINN_PATH must point to it.

func Example_basic() {

	// Find sqinn
	sqinnPath := os.Getenv("SQINN_PATH")
	if sqinnPath == "" {
		log.Printf("SQINN_PATH not set, please install sqinn and set SQINN_PATH")
		return
	}

	// Launch sqinn. Terminate at program exit.
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	defer sq.Terminate()

	// Open database. Close when we're done.
	sq.MustOpen(":memory:")
	defer sq.Close()

	// Create a table.
	sq.MustExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users.
	sq.MustExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	sq.MustExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

	// Query users.
	rows := sq.MustQuery("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
	for _, row := range rows {
		fmt.Printf("%d %q\n", row.Values[0].AsInt(), row.Values[1].AsString())
	}
	// Output:
	// 1 "Alice"
	// 2 "Bob"
}

func Example_parameterBinding() {
	// Find sqinn
	sqinnPath := os.Getenv("SQINN_PATH")
	if sqinnPath == "" {
		log.Printf("SQINN_PATH not set, please install sqinn and set SQINN_PATH")
		return
	}

	// Launch sqinn.
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	defer sq.Terminate()

	// Open database.
	sq.MustOpen(":memory:")
	defer sq.Close()

	// Create table
	sq.MustExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert 3 rows at once
	sq.MustExecOne("BEGIN")
	sq.MustExec(
		"INSERT INTO users (id, name) VALUES (?,?)",
		3, // insert 3 rows
		2, // each row has 2 columns
		[]any{
			1, "Alice", // bind first row
			2, "Bob", // bind second row
			3, nil, // third row has no name
		},
	)
	sq.MustExecOne("COMMIT")

	// Query rows
	rows := sq.MustQuery(
		"SELECT id, name FROM users WHERE id < ? ORDER BY id ASC",
		[]any{42},                           // WHERE id < 42
		[]byte{sqinn.ValInt, sqinn.ValText}, // two columns: int id, string name
	)
	for _, row := range rows {
		fmt.Printf("%d %q\n", row.Values[0].AsInt(), row.Values[1].AsString())
	}
	// Output:
	// 1 "Alice"
	// 2 "Bob"
	// 3 ""
}

func Example_handlingNullValues() {
	// Launch sqinn. Env SQINN_PATH must point to sqinn binary.
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	defer sq.Terminate()

	// Open database.
	sq.MustOpen(":memory:")
	defer sq.Close()

	// Create table
	sq.MustExecOne("CREATE TABLE names (val TEXT)")

	// Insert 2 rows, the first is non-NULL, the second is NULL
	sq.MustExecOne("BEGIN")
	sq.MustExec(
		"INSERT INTO names (val) VALUES (?)",
		2, // insert 2 rows
		1, // each row has 1 column
		[]any{
			"wombat", // first row is 'wombat'
			nil,      // second row is NULL
		},
	)
	sq.MustExecOne("COMMIT")

	// Query rows
	rows := sq.MustQuery(
		"SELECT val FROM names ORDER BY val",
		nil,                   // no query parameters
		[]byte{sqinn.ValText}, // one column of type TEXT
	)
	for _, row := range rows {
		stringValue := row.Values[0].String
		if stringValue.IsNull() {
			fmt.Printf("NULL\n")
		} else {
			fmt.Printf("%q\n", stringValue.Value)
		}
	}
	// Output:
	// NULL
	// "wombat"
}

func Example_sqliteSpecialties() {
	// Launch sqinn. Env SQINN_PATH must point to sqinn binary.
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	defer sq.Terminate()

	// Open database.
	sq.MustOpen(":memory:")
	defer sq.Close()

	// Enable foreign keys, see https://sqlite.org/pragma.html#pragma_foreign_keys
	sq.MustExecOne("PRAGMA foreign_keys = 1")

	// Set busy_timeout, see https://sqlite.org/pragma.html#pragma_busy_timeout
	sq.MustExecOne("PRAGMA busy_timeout = 10000")

	// Enable WAL mode, see https://sqlite.org/pragma.html#pragma_journal_mode
	sq.MustExecOne("PRAGMA journal_mode = WAL")

	// Enable NORMAL sync, see https://sqlite.org/pragma.html#pragma_synchronous
	sq.MustExecOne("PRAGMA synchronous = NORMAL")

	// Make a backup into a temp file
	filename := filepath.Join(os.TempDir(), "db_backup.sqlite")
	os.Remove(filename) // remove in case it exists, sqlite does not want to overwrite
	sq.MustExec("VACUUM INTO ?", 1, 1, []any{
		filename,
	})

	// Output:
}
