/*
A simple usage demo.
*/
package main

import (
	"log"
	"os"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// Simple sqinn-go usage. Error handling is left out. Parameter binding is also left out.
func main() {

	// Launch sqinn. Sqinn executable path is taken from environment.
	sq, _ := sqinn.New(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})

	// Open a database. Database file will be created if it does not exist.
	sq.Open("./users.db")

	// Create a table.
	sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users.
	sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
	sq.ExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

	// Query users. Two columns: id (int) and name (string)
	rows, _ := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
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

// Launch sqinn with path
func launchWithPath() {
	sq, _ := sqinn.New(sqinn.Options{
		SqinnPath: "C:/projects/my_server/bin/sqinn.exe",
	})
	_ = sq
}

// Use logger
func launchWithLogger() {
	sq, _ := sqinn.New(sqinn.Options{
		Logger: sqinn.StdLogger{},
	})
	_ = sq
}

// Insert three users.
func insertUsers(sq *sqinn.Sqinn) {
	sq.ExecOne("BEGIN")
	nUsers := 3
	nParamsPerUser := 2
	sq.Exec("INSERT INTO users (id, name) VALUES (?, ?)", nUsers, nParamsPerUser, []interface{}{
		1, "Alice",
		2, "Bob",
		3, nil,
	})
	sq.ExecOne("COMMIT")
}

// Query users where id < 42.
func queryUsers(sq *sqinn.Sqinn) ([]sqinn.Row, error) {
	rows, err := sq.Query(
		"SELECT id, name FROM users WHERE id < ? ORDER BY name",
		[]interface{}{42},                   // WHERE id < 42
		[]byte{sqinn.ValInt, sqinn.ValText}, // two columns: int id and string name
	)
	return rows, err
}
