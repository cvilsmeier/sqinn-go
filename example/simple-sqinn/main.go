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
