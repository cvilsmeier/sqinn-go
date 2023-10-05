package sqinn_test

import (
	"os"
	"strings"
	"testing"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// NOTE: For running the tests you must have
// sqinn(.exe) binary installed
// and $SQINN_PATH must point to it.
// You can download pre-built sqinn binaries from
// https://github.com/cvilsmeier/sqinn/releases

func TestOpenAndClose(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, sq != nil, "want sq but was nil")
	// get versions
	str, err := sq.SqinnVersion()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, str != "", "wrong sqinn version %q", str)
	io, err := sq.IoVersion()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, io == 1, "wrong io version %v", io)
	str, err = sq.SqliteVersion()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, str != "", "wrong sqlite version %q", str)
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	// open memory db again, must fail
	err = sq.Open(":memory:")
	assert(t, err != nil, "want err but was ok")
	substr := "already open"
	assert(t, strings.Contains(err.Error(), substr), "want ..%q.. but was %s", substr, err)
	// create table
	_, err = sq.ExecOne("CREATE TABLE users (name VARCHAR)")
	assert(t, err == nil, "want ok but was %s", err)
	// table must be empty
	rows := sq.MustQuery("SELECT name FROM users", nil, []byte{sqinn.ValText})
	assert(t, len(rows) == 0, "want no rows but was %d", len(rows))
	// close db
	err = sq.Close()
	assert(t, err == nil, "want ok but was %s", err)
	// open memory db #2
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	// query users must fail because table does not exist in db #2
	_, err = sq.Query("SELECT name FROM users ORDER BY name", nil, []byte{sqinn.ValText})
	assert(t, err != nil, "want err but was ok")
	substr = "no such table"
	assert(t, strings.Contains(err.Error(), substr), "want ..%q.. but was %s", substr, err)
	// close db #2
	err = sq.Close()
	assert(t, err == nil, "want ok but was %s", err)
	// close db #2 again must work
	err = sq.Close()
	assert(t, err == nil, "want ok but was %s", err)
	// terminate sqinn
	err = sq.Terminate()
	assert(t, err == nil, "want ok but was %s", err)
}

func TestMustExecQuery(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// exec must work
	sq.MustExec("DROP TABLE IF EXISTS users", 1, 0, nil)
	sq.MustExecOne("DROP TABLE IF EXISTS users")
	sq.MustQuery("SELECT 1", nil, nil)
}

func TestExecAndQueryWithErrors(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// create a table
	sq.MustExecOne("CREATE TABLE users (name VARCHAR NOT NULL)")
	// insert user with wrong sql, must fail
	_, err = sq.ExecOne("INSERT INTO users (id, name, age) VALUES (1, 'Alice', 27)")
	assert(t, err != nil, "want err but was ok")
	// insert user with wrong param (NOT NULL!!), must fail
	_, err = sq.Exec("INSERT INTO users (name) VALUES (?)", 1, 1, []any{nil})
	assert(t, err != nil, "want err but was ok")
	// insert user with good sql, must succeed
	_, err = sq.ExecOne("INSERT INTO users (name) VALUES ('Alice')")
	assert(t, err == nil, "want ok but was %v", err)
	// query users with wrong sql, must fail
	rows, err := sq.Query("SELECT id, name, age FROM users ORDER BY id", nil, nil)
	assert(t, err != nil, "want err but was ok")
	assert(t, len(rows) == 0, "want no rows but was %v", len(rows))
	// query users with good sql, must succeed
	rows, err = sq.Query("SELECT name FROM users ORDER BY name", nil, []byte{sqinn.ValText})
	assert(t, err == nil, "want ok but was %v", err)
	assert(t, len(rows) == 1, "want 1 row but was %v", len(rows))
}

func TestColTypes(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// create table with all possible types
	_, err = sq.ExecOne("CREATE TABLE foo (i INTEGER, i64 BIGINT, f64 REAL, s TEXT, b BLOB)")
	assert(t, err == nil, "want ok but was %s", err)
	// insert row with values
	mods, err := sq.Exec(
		"INSERT INTO foo (i, i64, f64, s, b) VALUES(?, ?, ?, ?, ?)", // sql
		1, // 1 row
		5, // row has 5 columns
		[]any{
			13,              // int i
			int64(1) << 62,  // int64 i64
			float64(1.002),  // float64 f64
			"partes tres",   // string s
			[]byte{1, 2, 3}, // blob b
		},
	)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, len(mods) == 1, "want 1 mods but was %d", len(mods))
	assert(t, mods[0] == 1, "want mod 1 but was %d", mods[0])
	// insert row with all NULLs
	mod, err := sq.ExecOne("INSERT INTO foo (i, i64, f64, s, b) VALUES(NULL, NULL, NULL, NULL, NULL)")
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, mod == 1, "want mod 1 but was %d", mod)
	// query all rows
	rows, err := sq.Query(
		"SELECT i, i64, f64, s, b FROM foo ORDER BY i",
		nil, // no query parameters
		[]byte{
			sqinn.ValInt,
			sqinn.ValInt64,
			sqinn.ValDouble,
			sqinn.ValText,
			sqinn.ValBlob,
		}, // 5 column types
	)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, len(rows) == 2, "want 2 rows but was %d", len(rows))
	values := rows[0].Values
	assert(t, len(values) == 5, "want 5 values but was %d", len(values))
	assert(t, !values[0].Int.Set, "want i NULL")
	assert(t, !values[1].Int64.Set, "want i64 NULL")
	assert(t, !values[2].Double.Set, "want f64 NULL")
	assert(t, !values[3].String.Set, "want s NULL")
	assert(t, !values[4].Blob.Set, "want b NULL")
	values = rows[1].Values
	assert(t, len(values) == 5, "want 5 values but was %d", len(values))
	assert(t, values[0].Int.Set, "want int set")
	assert(t, values[0].Int.Value == 13, "wrong value")
	assert(t, values[1].Int64.Set, "want int64 set")
	assert(t, values[1].Int64.Value == 4611686018427387904, "wrong value")
	assert(t, 0x4000000000000000 == 4611686018427387904, "wrong value")
	assert(t, int64(1)<<62 == 4611686018427387904, "wrong value")
	assert(t, 1<<62 == 4611686018427387904, "wrong value")
	assert(t, values[2].Double.Set, "want double set")
	assert(t, values[2].Double.Value == 1.002, "wrong value")
	assert(t, values[3].String.Set, "want string set")
	assert(t, values[3].String.Value == "partes tres", "wrong value")
	assert(t, values[4].Blob.Set, "want blob set")
	assert(t, len(values[4].Blob.Value) == 3, "want blob len 3 but was %d", len(values[4].Blob.Value))
	assert(t, values[4].Blob.Value[0] == 1, "want blob[0] 1 but was %d", values[4].Blob.Value[0])
	assert(t, values[4].Blob.Value[1] == 2, "want blob[1] 2 but was %d", values[4].Blob.Value[1])
	assert(t, values[4].Blob.Value[2] == 3, "want blob[2] 3 but was %d", values[4].Blob.Value[2])
	// delete all rows
	mod, err = sq.ExecOne("DELETE FROM foo")
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, mod == 2, "want mod 2 but was %d", mod)
	// query all rows again, must have none
	rows, err = sq.Query(
		"SELECT i, i64, f64, s, b FROM foo ORDER BY i",
		nil,                  // no query parameters
		[]byte{sqinn.ValInt}, // we want only the first column
	)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, len(rows) == 0, "want len(rows) 0 but was %d", len(rows))
}

func TestNullValues(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// create table with all possible types
	_, err = sq.ExecOne("CREATE TABLE tabl (i INTEGER, i64 BIGINT, f64 REAL, s TEXT, b BLOB)")
	assert(t, err == nil, "want ok but was %s", err)
	// insert row with NULL values
	mods, err := sq.Exec(
		"INSERT INTO tabl (i, i64, f64, s, b) VALUES(?, ?, ?, ?, ?)", // sql
		1, // insert 1 row
		5, // row has 5 columns
		[]any{
			nil, // int i
			nil, // int64 i64
			nil, // float64 f64
			nil, // string s
			nil, // blob b
		},
	)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, len(mods) == 1, "want 1 mods but was %d", len(mods))
	assert(t, mods[0] == 1, "want mod 1 but was %d", mods[0])
	// query all rows
	rows, err := sq.Query(
		"SELECT i, i64, f64, s, b FROM tabl ORDER BY i",
		nil, // no query parameters
		[]byte{
			sqinn.ValInt,    // query 'i INTEGER' as int
			sqinn.ValInt64,  // query 'i64 BIGINT' as int64
			sqinn.ValDouble, // query 'f64 REAL' as float64
			sqinn.ValText,   // query 's TEXT' as string
			sqinn.ValBlob,   // query 'b BLOB' as []byte
		}, // 5 column types
	)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, len(rows) == 1, "want 1 row but was %d", len(rows))
	values := rows[0].Values
	assert(t, len(values) == 5, "want 5 values but was %d", len(values))
	assert(t, !values[0].Int.Set, "want i NULL")
	assert(t, !values[1].Int64.Set, "want i64 NULL")
	assert(t, !values[2].Double.Set, "want f64 NULL")
	assert(t, !values[3].String.Set, "want s NULL")
	assert(t, !values[4].Blob.Set, "want b NULL")
}

func TestLaunchError(t *testing.T) {
	// launching a non-existing binary must not work
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: "./does_surely_not_exist/sqinn.exe",
	})
	assert(t, sq == nil, "want sq == nil but was set")
	assert(t, err != nil, "want err but was ok")
	substr := "does_surely_not_exist"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
}

func TestLowLevelFunctions(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// create schema
	err = sq.Prepare("CREATE TABLE users (name VARCHAR)")
	assert(t, err == nil, "want ok but was %s", err)
	more, err := sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, !more, "want !more but was %v", more)
	err = sq.Finalize()
	assert(t, err == nil, "want ok but was %s", err)
	// insert two users Alice and Bob
	err = sq.Prepare("INSERT INTO users (name) VALUES (?)")
	assert(t, err == nil, "want ok but was %s", err)
	err = sq.Bind(1, "Alice")
	assert(t, err == nil, "want ok but was %s", err)
	_, err = sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	mod, err := sq.Changes()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, mod == 1, "want mod == 1 but was %v", mod)
	err = sq.Reset()
	assert(t, err == nil, "want ok but was %s", err)
	err = sq.Bind(1, "Bob")
	assert(t, err == nil, "want ok but was %s", err)
	_, err = sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	mod, err = sq.Changes()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, mod == 1, "want mod == 1 but was %v", mod)
	err = sq.Finalize()
	assert(t, err == nil, "want ok but was %s", err)
	// query users table
	err = sq.Prepare("SELECT name FROM users WHERE name <> ?")
	assert(t, err == nil, "want ok but was %s", err)
	err = sq.Bind(1, "Bob")
	assert(t, err == nil, "want ok but was %s", err)
	more, err = sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, more, "want more but was %v", more)
	any, err := sq.Column(0, sqinn.ValText)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, any.String.Set, "want any.String.Set but was %v", any.String.Set)
	assert(t, any.String.Value == "Alice", "want any.String.Value == 'Alice' but was %v", any.String.Value)
	err = sq.Reset()
	assert(t, err == nil, "want ok but was %s", err)
	err = sq.Bind(1, "Alice")
	assert(t, err == nil, "want ok but was %s", err)
	more, err = sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, more, "want more but was %v", more)
	any, err = sq.Column(0, sqinn.ValText)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, any.String.Set, "want any.String.Set but was %v", any.String.Set)
	assert(t, any.String.Value == "Bob", "want any.String.Value == 'Bob' but was %v", any.String.Value)
	err = sq.Finalize()
	assert(t, err == nil, "want ok but was %s", err)
}

func TestMisuse(t *testing.T) {
	// launch
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: os.Getenv("SQINN_PATH"),
	})
	assert(t, sq != nil, "no sq")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Terminate()
	// open db
	err = sq.Open(":memory:")
	assert(t, err == nil, "want ok but was %s", err)
	defer sq.Close()
	// double open must fail
	err = sq.Open(":memory:")
	assert(t, err != nil, "want err but was ok")
	substr := "already open"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// prepare statement
	err = sq.Prepare("PRAGMA foreign_keys;")
	assert(t, err == nil, "want ok but was %s", err)
	// double prepare must fail
	err = sq.Prepare("PRAGMA foreign_keys;")
	assert(t, err != nil, "want err but was ok")
	substr = "must finalize first"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// bind non-existent parameter must fail
	err = sq.Bind(1, "bind_me")
	assert(t, err != nil, "want err but was ok")
	substr = "column index out of range"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// bind iparam < 1 must fail
	err = sq.Bind(0, "bind_me")
	assert(t, err != nil, "want err but was ok")
	substr = "iparam must be >= 1"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// step must work
	more, err := sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, more, "want more but was %t", more)
	// column 0 must work
	any, err := sq.Column(0, sqinn.ValInt)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, any.Int.Set, "want any.Int.Set but was %t", any.Int.Set)
	// column 1 must work but must be NULL
	any, err = sq.Column(1, sqinn.ValInt)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, !any.Int.Set, "want !any.Int.Set but was %t", any.Int.Set)
	// column 113 must work but must be NULL
	any, err = sq.Column(113, sqinn.ValInt)
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, !any.Int.Set, "want !any.Int.Set but was %t", any.Int.Set)
	// step must work but have no more
	more, err = sq.Step()
	assert(t, err == nil, "want ok but was %s", err)
	assert(t, !more, "want !more but was %t", more)
	// reset must work
	err = sq.Reset()
	assert(t, err == nil, "want ok but was %s", err)
	// double reset must work
	err = sq.Reset()
	assert(t, err == nil, "want ok but was %s", err)
	// close must fail as long as we have active statements
	err = sq.Close()
	assert(t, err != nil, "want err but was ok")
	substr = "unable to close due to unfinalized statements"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// exec must fail as long as we have active statements
	_, err = sq.Exec("SELECT 1", 1, 0, nil)
	assert(t, err != nil, "want err but was ok")
	substr = "must finalize first"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// query must fail as long as we have active statements
	_, err = sq.Query("SELECT 1", nil, []byte{sqinn.ValInt})
	assert(t, err != nil, "want err but was ok")
	substr = "must finalize first"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// finalize must work
	err = sq.Finalize()
	assert(t, err == nil, "want ok but was %s", err)
	// double finalize must work
	err = sq.Finalize()
	assert(t, err == nil, "want ok but was %s", err)
	// bind non-existent parameter must fail with a misleading error message
	err = sq.Bind(1, "bind_me")
	assert(t, err != nil, "want err but was ok")
	substr = "unable to close due to unfinalized statements"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// step must fail with a misleading error message
	_, err = sq.Step()
	assert(t, err != nil, "want err but was ok")
	substr = "unable to close due to unfinalized statements"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// close must work
	err = sq.Close()
	assert(t, err == nil, "want ok but was %s", err)
	// double close must work
	err = sq.Close()
	assert(t, err == nil, "want ok but was %s", err)
	// step must fail with a funny error message
	_, err = sq.Step()
	assert(t, err != nil, "want err but was ok")
	substr = "out of memory"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
	// bind must fail with a funny error message
	err = sq.Bind(42, 13)
	assert(t, err != nil, "want err but was ok")
	substr = "out of memory"
	assert(t, strings.Contains(err.Error(), substr), "want %q but was %s", substr, err)
}

func assert(t testing.TB, cond bool, format string, args ...any) {
	t.Helper()
	if !cond {
		t.Fatalf(format, args...)
	}
}

func BenchmarkValueBinding(b *testing.B) {
	bindFunc := func(value any) byte {
		switch value.(type) {
		case nil:
			return sqinn.ValNull
		case int:
			return sqinn.ValInt
		case int64:
			return sqinn.ValInt64
		case float64:
			return sqinn.ValDouble
		case string:
			return sqinn.ValText
		case []byte:
			return sqinn.ValBlob
		}
		return sqinn.ValNull
	}
	intValue := int(1)
	int64Value := int64(1)
	float64Value := float64(1)
	stringValue := "1"
	blobValue := []byte{1}
	values := []any{nil, intValue, int64Value, float64Value, stringValue, blobValue}
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			valType := bindFunc(value)
			_ = valType
		}
	}
}

func BenchmarkValueBindingWithPointers(b *testing.B) {
	bindFunc := func(value any) byte {
		switch value.(type) {
		case nil:
			return sqinn.ValNull
		case int:
			return sqinn.ValInt
		case *int:
			return sqinn.ValInt
		case int64:
			return sqinn.ValInt64
		case *int64:
			return sqinn.ValInt
		case float64:
			return sqinn.ValDouble
		case *float64:
			return sqinn.ValDouble
		case string:
			return sqinn.ValText
		case *string:
			return sqinn.ValText
		case []byte:
			return sqinn.ValBlob
		}
		return sqinn.ValNull
	}
	intValue := int(1)
	int64Value := int64(1)
	float64Value := float64(1)
	stringValue := "1"
	blobValue := []byte{1}
	values := []any{nil, intValue, &intValue, int64Value, &int64Value, float64Value, &float64Value, stringValue, &stringValue, blobValue}
	for i := 0; i < b.N; i++ {
		for _, value := range values {
			valType := bindFunc(value)
			_ = valType
		}
	}
}
