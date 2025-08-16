/*
A simple usage demo for sqinn-go.
*/
package main

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
	//
	// Create a table, cleanup when done
	sq.MustExecSql("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name TEXT)")
	defer sq.MustExecSql("DROP TABLE users")
	//
	// Insert users
	sq.MustExecParams("INSERT INTO users (id, name) VALUES (?, ?)", 3, 2, []sqinn.Value{
		sqinn.Int32Value(1), sqinn.StringValue("Alice"),
		sqinn.Int32Value(2), sqinn.StringValue("Bob"),
		sqinn.Int32Value(3), sqinn.StringValue("Carol"),
	})
	//
	// Query users
	rows := sq.MustQueryRows(
		"SELECT id, name FROM users WHERE id >= ? ORDER BY id",
		[]sqinn.Value{sqinn.Int32Value(0)},      // query parameters
		[]byte{sqinn.ValInt32, sqinn.ValString}, // fetch id as int, name as string
	)
	for _, values := range rows {
		fmt.Printf("user id=%d, name=%s\n", values[0].Int32, values[1].String)
	}
	//
	// Output:
	// user id=1, name=Alice
	// user id=2, name=Bob
	// user id=3, name=Carol
}
