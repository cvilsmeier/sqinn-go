/*
A simple usage demo.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// Simple sqinn-go usage. Error handling left out for brevity.
func main() {
	sqinnpath := os.Getenv("SQINN_PATH")
	dbname := "./users.db"
	flag.StringVar(&sqinnpath, "sqinn", sqinnpath, "path to sqinn")
	flag.StringVar(&dbname, "db", dbname, "path to db file")

	// Launch sqinn.
	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnpath,
	})
	// Terminate when program exists.
	defer sq.Terminate()

	// Open a database. Database file will be created if it does not exist.
	sq.Open(dbname)
	// Close when done.
	defer sq.Close()

	// Create a table.
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert user without parameters.
	sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")

	// Insert three users in a transaction.
	sq.ExecOne("BEGIN")
	nusers := 3         // we want three users
	nparamsPerUser := 2 // each user has 2 columns: id and name
	paramValues := []interface{}{
		2, "Bob", // values for first user
		3, "Carol", // values for second user
		4, "Dave", // values for third user
	}
	sq.Exec("INSERT INTO users (id, name) VALUES (?, ?)", nusers, nparamsPerUser, paramValues)
	sq.ExecOne("COMMIT")

	// Query all users. Two columns: id (int) and name (string)
	rows, _ := sq.Query(
		"SELECT id, name FROM users ORDER BY id",
		nil,                                 // no query parameters
		[]byte{sqinn.ValInt, sqinn.ValText}, // fetch id as int, name as string
	)
	for _, row := range rows {
		fmt.Printf("found id=%d, name=%s\n", row.Values[0].AsInt(), row.Values[1].AsString())
	}
	// output:
	// found id=1, name=Alice
	// found id=2, name=Bob
	// found id=3, name=Carol
	// found id=4, name=Dave

	// Query name for id 2
	id := 2
	rows, _ = sq.Query(
		"SELECT name FROM users WHERE id = ?",
		[]interface{}{id},     // WHERE id = 2
		[]byte{sqinn.ValText}, // fetch name as string
	)
	for _, row := range rows {
		fmt.Printf("id %d is %q\n", id, row.Values[0].AsString())
	}
	// output:
	// id 2 is "Bob"

	// Delete users.
	modCount, _ := sq.ExecOne("DELETE FROM users")
	fmt.Printf("deleted %d rows\n", modCount)
	// output:
	// deleted 4 rows

	// Cleanup.
	sq.ExecOne("DROP TABLE users")
}

// Launch sqinn with path
func launchWithPath() {
	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: "C:/projects/my_server/bin/sqinn.exe",
	})
	_ = sq
}
