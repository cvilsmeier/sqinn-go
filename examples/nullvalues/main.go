/*
Demonstrates NULL handling for sqinn
*/
package main

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
	sq.MustExecSql("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name TEXT)")
	defer sq.MustExecSql("DROP TABLE users")

	// Insert users, some with NULL values.
	sq.MustExec("INSERT INTO users (id, name) VALUES (?, ?)", [][]any{
		{1, "Alice"},
		{2, nil},
		{3, nil},
	})

	// Query users, be aware that name can be NULL.
	rows := sq.MustQuery(
		"SELECT id, name FROM users WHERE id >= ? ORDER BY id",
		[]any{0},                                // query parameters
		[]byte{sqinn.ValInt32, sqinn.ValString}, // fetch id as int, name as string
	)
	for _, values := range rows {
		id := values[0].Int32
		nameIsNull := values[1].Type == sqinn.ValNull
		if nameIsNull {
			fmt.Printf("user id=%d, name=NULL\n", id)
		} else {
			name := values[1].String
			fmt.Printf("user id=%d, name=%s\n", id, name)
		}
	}

	// Output:
	// user id=1, name=Alice
	// user id=2, name=NULL
	// user id=3, name=NULL
}
