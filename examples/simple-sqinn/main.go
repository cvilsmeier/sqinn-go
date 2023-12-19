/*
A simple usage demo for sqinn-go.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// Simple sqinn-go usage.
func main() {
	sqinnpath := os.Getenv("SQINN_PATH")
	dbname := ":memory:" // or a real file, e.g. "/tmp/users.db"
	flag.StringVar(&sqinnpath, "sqinn", sqinnpath, "path to sqinn")
	flag.StringVar(&dbname, "db", dbname, "path to db file")
	flag.Parse()

	// Launch sqinn, terminate when program exists.
	sq := sqinn.MustLaunch(sqinn.Options{
		SqinnPath: sqinnpath,
	})
	defer sq.Terminate()

	// Open a database. Database file will be created if it
	// does not exist. Close when done.
	sq.MustOpen(dbname)
	defer sq.Close()

	// Create a table. Cleanup when done.
	sq.MustExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")
	defer sq.ExecOne("DROP TABLE users")

	// Insert user without parameters.
	sq.MustExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")

	// Insert three users in a transaction.
	sq.MustExecOne("BEGIN")
	nusers := 3         // we want three users
	nparamsPerUser := 2 // each user has 2 columns: id and name
	paramValues := []any{
		2, "Bob", // values for first user
		3, "Carol", // values for second user
		4, "Dave", // values for third user
	}
	sq.MustExec("INSERT INTO users (id, name) VALUES (?, ?)", nusers, nparamsPerUser, paramValues)
	sq.MustExecOne("COMMIT")

	// Query all users. Two columns: id (int) and name (string)
	rows := sq.MustQuery(
		"SELECT id, name FROM users ORDER BY id",
		nil, // no query parameters
		[]sqinn.ValueType{sqinn.ValInt, sqinn.ValText}, // fetch id as int, name as string
	)
	for _, row := range rows {
		fmt.Printf("found id=%d, name=%s\n", row.Values[0].AsInt(), row.Values[1].AsString())
	}
	// Output:
	// found id=1, name=Alice
	// found id=2, name=Bob
	// found id=3, name=Carol
	// found id=4, name=Dave

	// Query name for id 2
	id := 2
	rows = sq.MustQuery(
		"SELECT name FROM users WHERE id = ?",
		[]any{id},             // WHERE id = 2
		[]sqinn.ValueType{sqinn.ValText}, // fetch name as string
	)
	for _, row := range rows {
		fmt.Printf("id %d is %q\n", id, row.Values[0].AsString())
	}
	// Output:
	// id 2 is "Bob"

	// Delete users.
	modCount := sq.MustExecOne("DELETE FROM users")
	fmt.Printf("deleted %d rows\n", modCount)
	// Output:
	// deleted 4 rows
}
