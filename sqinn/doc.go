/*
Package sqinn provides a pure-Go interface to SQLite.

It uses Sqinn (http://github.com/cvilsmeier/sqinn) for accessing SQLite
databases.

Basic Usage

The following sample code opens a database, inserts some data, queries it,
and closes the database. Error handling is left out for brevity.

	import "github.com/cvilsmeier/sqinn-go/sqinn"

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

Parameter Binding

Parameter Binding is supported for Exec and Query:

	// Insert three users in one transaction.
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


Options

Sqinn searches the sqinn binary in the PATH environment. You can customize that behavior by
specifying the path to sqinn explicitely when launching sqinn.

	sq, _ := sqinn.New(sqinn.Options{
		SqinnPath: "C:/projects/my_server/bin/sqinn.exe",
	})

The sqinn subprocess prints debug and error messages on its stderr. You can
consume it by setting a sqinn.Logger when launching sqinn.

	sq, _ := sqinn.New(sqinn.Options{
		Logger: sqinn.StdLogger{},
	})

See the sqinn.Logger docs for more details.

*/
package sqinn
