/*
Package sqinn provides interface to SQLite databases in pure Go.

It uses Sqinn (http://github.com/cvilsmeier/sqinn) for accessing SQLite
databases. It is not a database/sql driver.


Basic Usage

The following sample code opens a database, inserts some data, queries it,
and closes the database. Error handling is left out for brevity.

	import "github.com/cvilsmeier/sqinn-go/sqinn"

	func main() {

		// Launch sqinn.
		sq, _ := sqinn.Launch(sqinn.Options{})
		defer sq.Terminate()

		// Open database.
		sq.Open("./users.db")
		defer sq.Close()

		// Create a table.
		sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")

		// Insert users.
		sq.ExecOne("INSERT INTO users (id, name) VALUES (1, 'Alice')")
		sq.ExecOne("INSERT INTO users (id, name) VALUES (2, 'Bob')")

		// Query users.
		rows, _ := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
		for _, row := range rows {
			fmt.Printf("id=%d, name=%s\n", row.Values[0].AsInt(), row.Values[1].AsString())
		}

		// Output:
		// id=1, name=Alice
		// id=2, name=Bob
	}


Parameter Binding

For Exec:

	nUsers := 3
	nParamsPerUser := 2
	sq.Exec("INSERT INTO users (id, name) VALUES (?, ?)", nUsers, nParamsPerUser, []interface{}{
		1, "Alice",
		2, "Bob",
		3, nil,
	})

For Query:

	// Query users where id < 42.
	rows, err := sq.Query(
		"SELECT id, name FROM users WHERE id < ? ORDER BY name",
		[]interface{}{42},                   // WHERE id < 42
		[]byte{sqinn.ValInt, sqinn.ValText}, // two columns: int id and string name
	)


Options

Sqinn searches the sqinn binary in the $PATH environment. You can customize
that behavior by specifying the path to sqinn explicitly when launching sqinn.

	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: "C:/projects/wombat/bin/sqinn.exe",
	})

	// or, even better

	sq, _ := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})

The sqinn subprocess prints debug and error messages on its stderr. You can
consume it by setting a sqinn.Logger.

	sq, _ := sqinn.Launch(sqinn.Options{
		Logger: sqinn.StdLogger{},
	})

See the sqinn.Logger docs for more details.


Low-level Functions

Sqinn implements many of SQLite's C API low-level functions prepare(),
finalize(), step(), etc. Although made available, we recommend not using them.
Instead, use Exec and Query. Most (if not all) database tasks can be
accomplished with Exec and Query.

*/
package sqinn
