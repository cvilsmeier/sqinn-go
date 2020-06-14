package sqinn_test

import (
	"fmt"
	"log"
	"os"

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
	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	defer sq.Terminate()

	// Open database. Close when we're done.
	sq.Open(":memory:")
	defer sq.Close()

	// Create a table.
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users.
	sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	sq.ExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

	// Query users.
	rows, _ := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
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
	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	defer sq.Terminate()

	// Open database.
	sq.Open(":memory:")
	defer sq.Close()

	// Create table
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert 3 rows at once
	sq.ExecOne("BEGIN")
	sq.Exec(
		"INSERT INTO users (id, name) VALUES (?,?)",
		3, // insert 3 rows
		2, // each row has 2 columns
		[]interface{}{
			1, "Alice", // bind first row
			2, "Bob", // bind second row
			3, nil, // third row has no name
		},
	)
	sq.ExecOne("COMMIT")

	// Query rows
	rows, _ := sq.Query(
		"SELECT id, name FROM users WHERE id < ? ORDER BY id ASC",
		[]interface{}{42},                   // WHERE id < 42
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
