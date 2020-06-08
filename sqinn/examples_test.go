package sqinn_test

import (
	"fmt"
	"log"
	"os"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func Example_basic() {
	// Launch sqinn, sqinn-path is taken from environment
	sq, err := sqinn.New(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
		Logger:    sqinn.StdLogger{},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sq.Terminate()

	// Open in-memory database
	err = sq.Open(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer sq.Close()

	// Create table
	sq.MustExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

	// Insert users
	sq.MustExecOne("BEGIN")
	sq.MustExec(
		"INSERT INTO users (id, name) VALUES (?,?)",
		3, // we have 3 users
		2, // each user 2 columns
		[]interface{}{
			1, "Alice",
			2, "Bob",
			3, nil,
		},
	)
	sq.MustExecOne("COMMIT")

	// Query users
	rows := sq.MustQuery(
		"SELECT id, name FROM users WHERE id < ? ORDER BY id ASC",
		[]interface{}{42},                   // WHERE id < 42
		[]byte{sqinn.ValInt, sqinn.ValText}, // two columns: int id, string name
	)
	for _, row := range rows {
		fmt.Printf(
			"%d '%s'\n",
			row.Values[0].AsInt(),
			row.Values[1].AsString(),
		)
	}

	// Output:
	// 1 'Alice'
	// 2 'Bob'
	// 3 ''
}
