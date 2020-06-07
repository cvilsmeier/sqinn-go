package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

// var SqinnPath = "sqinn"
// var DbFilename = "test.db"

func testFunctions(sqinnPath, dbFile string, nusers int) {
	funcname := "testFunctions"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d", sqinnPath, dbFile, nusers)
	assert := func(c bool) {
		if !c {
			panic("assertion failed")
		}
	}
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	sq, err := sqinn.NewSqinn(sqinnPath, sqinn.StdLogger{})
	check(err)
	// open db
	err = sq.Open(dbFile)
	check(err)
	t1 := time.Now()
	// prepare schema
	sql := "DROP TABLE IF EXISTS users"
	err = sq.Prepare(sql)
	check(err)
	more, err := sq.Step()
	check(err)
	assert(!more)
	err = sq.Finalize()
	check(err)
	sql = "CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	// insert users
	sql = "BEGIN TRANSACTION"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	sql = "INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)"
	err = sq.Prepare(sql)
	check(err)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + i
		rating := 0.13 * float64(i+1)
		check(sq.Bind(1, id))
		check(sq.Bind(2, name))
		check(sq.Bind(3, age))
		check(sq.Bind(4, rating))
		_, err = sq.Step()
		check(err)
		check(sq.Reset())
		ch, err := sq.Changes()
		check(err)
		assert(ch == 1)
	}
	err = sq.Finalize()
	check(err)
	sql = "COMMIT"
	err = sq.Prepare(sql)
	check(err)
	_, err = sq.Step()
	check(err)
	err = sq.Finalize()
	check(err)
	t2 := time.Now()
	// query users
	sql = "SELECT id, name, age, rating FROM users ORDER BY id"
	err = sq.Prepare(sql)
	check(err)
	more, err = sq.Step()
	check(err)
	var nrows int
	for more {
		nrows++
		idValue, err := sq.ColumnInt(0)
		check(err)
		nameValue, err := sq.ColumnText(1)
		check(err)
		ageValue, err := sq.ColumnInt(2)
		check(err)
		ratingValue, err := sq.ColumnDouble(3)
		check(err)
		_, _, _, _ = idValue, nameValue, ageValue, ratingValue
		// log.Printf("%d | %s | %d | %g", idValue.Value, nameValue.Value, ageValue.Value, ratingValue.Value)
		more, err = sq.Step()
		check(err)
	}
	log.Printf("fetched %d rows", nrows)
	err = sq.Finalize()
	check(err)
	t3 := time.Now()
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testUsers(sqinnPath, dbFile string, nusers int, bindRating bool) {
	funcname := "testUsers"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d, bindRating=%t", sqinnPath, dbFile, nusers, bindRating)
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	sq, err := sqinn.NewSqinn(sqinnPath, sqinn.StdLogger{})
	check(err)
	// open db
	err = sq.Open(dbFile)
	check(err)
	// prepare schema
	err = sq.Exec("DROP TABLE IF EXISTS users", 1, 0, nil)
	err = sq.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)", 1, 0, nil)
	// insert users
	t1 := time.Now()
	err = sq.Exec("BEGIN TRANSACTION", 1, 0, nil)
	values := make([]interface{}, 0, nusers*4)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + i
		rating := 0.13 * float64(i+1)
		if bindRating {
			values = append(values, id, name, age, rating)
		} else {
			values = append(values, id, name, age, nil)
		}
	}
	err = sq.Exec("INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)", nusers, 4, values)
	err = sq.Exec("COMMIT", 1, 0, nil)
	t2 := time.Now()
	// query users
	colTypes := []byte{sqinn.VAL_INT, sqinn.VAL_TEXT, sqinn.VAL_INT, sqinn.VAL_DOUBLE}
	rows, err := sq.Query("SELECT id, name, age, rating FROM users ORDER BY id", nil, colTypes)
	check(err)
	log.Printf("fetched %d rows", len(rows))
	// for _, row := range rows {
	// 	log.Printf("%d | %s | %d | %g",
	// 		row.Values[0].IntValue.Value,
	// 		row.Values[1].StringValue.Value,
	// 		row.Values[2].IntValue.Value,
	// 		row.Values[3].DoubleValue.Value,
	// 	)
	// }
	t3 := time.Now()
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testComplex(sqinnPath, dbFile string, nprofiles, nusers, nlocations int) {
	funcname := "testComplex"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nprofiles, nusers, nlocations = %d, %d, %d", sqinnPath, dbFile, nprofiles, nusers, nlocations)
	check := func(err error) {
		if err != nil {
			log.Fatal(err)
		}
	}
	sq, err := sqinn.NewSqinn(sqinnPath, sqinn.StdLogger{})
	check(err)
	// open db
	check(sq.Open(dbFile))
	check(sq.Exec("PRAGMA foreign_keys=1", 1, 0, nil))
	check(sq.Exec("DROP TABLE IF EXISTS locations", 1, 0, nil))
	check(sq.Exec("DROP TABLE IF EXISTS users", 1, 0, nil))
	check(sq.Exec("DROP TABLE IF EXISTS profiles", 1, 0, nil))
	check(sq.Exec("CREATE TABLE profiles (id VARCHAR PRIMARY KEY NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL)", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_profiles_name ON profiles(name);", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_profiles_active ON profiles(active);", 1, 0, nil))
	check(sq.Exec("CREATE TABLE users (id VARCHAR PRIMARY KEY NOT NULL, profileId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (profileId) REFERENCES profiles(id))", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_users_profileId ON users(profileId);", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_users_name ON users(name);", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_users_active ON users(active);", 1, 0, nil))
	check(sq.Exec("CREATE TABLE locations (id VARCHAR PRIMARY KEY NOT NULL, userId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (userId) REFERENCES users(id))", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_locations_userId ON locations(userId);", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_locations_name ON locations(name);", 1, 0, nil))
	check(sq.Exec("CREATE INDEX idx_locations_active ON locations(active);", 1, 0, nil))
	// insert
	t1 := time.Now()
	check(sq.Exec("BEGIN TRANSACTION", 1, 0, nil))
	values := make([]interface{}, 0, nprofiles*3)
	for p := 0; p < nprofiles; p++ {
		profileId := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("ProfileGo %d", p)
		active := p % 2
		values = append(values, profileId, name, active)
	}
	check(sq.Exec("INSERT INTO profiles (id,name,active) VALUES(?,?,?)", nprofiles, 3, values))
	check(sq.Exec("COMMIT", 1, 0, nil))
	check(sq.Exec("BEGIN TRANSACTION", 1, 0, nil))
	values = make([]interface{}, 0, nprofiles*nusers*4)
	for p := 0; p < nprofiles; p++ {
		profileId := fmt.Sprintf("profile_%d", p)
		for u := 0; u < nusers; u++ {
			userId := fmt.Sprintf("user_%d_%d", p, u)
			name := fmt.Sprintf("User %d %d", p, u)
			active := u % 2
			values = append(values, userId, profileId, name, active)
		}
	}
	check(sq.Exec("INSERT INTO users (id,profileId,name,active) VALUES(?,?,?,?)", nprofiles*nusers, 4, values))
	check(sq.Exec("COMMIT", 1, 0, nil))
	check(sq.Exec("BEGIN TRANSACTION", 1, 0, nil))
	values = make([]interface{}, 0, nprofiles*nusers*nlocations*4)
	for p := 0; p < nprofiles; p++ {
		for u := 0; u < nusers; u++ {
			userId := fmt.Sprintf("user_%d_%d", p, u)
			for l := 0; l < nlocations; l++ {
				locationId := fmt.Sprintf("location_%d_%d_%d", p, u, l)
				name := fmt.Sprintf("Location %d %d %d", p, u, l)
				active := l % 2
				values = append(values, locationId, userId, name, active)
			}
		}
	}
	check(sq.Exec("INSERT INTO locations (id,userId,name,active) VALUES(?,?,?,?)", nprofiles*nusers*nlocations, 4, values))
	check(sq.Exec("COMMIT", 1, 0, nil))
	t2 := time.Now()
	// query
	sql := "SELECT locations.id, locations.userId, locations.name, locations.active, users.id, users.profileId, users.name, users.active, profiles.id, profiles.name, profiles.active " +
		"FROM locations " +
		"LEFT JOIN users ON users.id = locations.userId " +
		"LEFT JOIN profiles ON profiles.id = users.profileId " +
		"WHERE locations.active = ? OR locations.active = ? " +
		"ORDER BY locations.name, locations.id, users.name, users.id, profiles.name, profiles.id"
	rows, err := sq.Query(sql, []interface{}{0, 1}, []byte{sqinn.VAL_TEXT, sqinn.VAL_TEXT, sqinn.VAL_TEXT, sqinn.VAL_INT, sqinn.VAL_TEXT, sqinn.VAL_TEXT, sqinn.VAL_TEXT, sqinn.VAL_INT, sqinn.VAL_TEXT, sqinn.VAL_TEXT, sqinn.VAL_INT})
	check(err)
	log.Printf("fetched %d rows", len(rows))
	t3 := time.Now()
	// close and terminate
	check(sq.Close())
	check(sq.Terminate())
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testBlob(sqinnPath, dbFile string) {
	funcname := "testBlob"
	log.Printf("TEST %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s", sqinnPath, dbFile)
	assert := func(c bool, format string, v ...interface{}) {
		if !c {
			panic(fmt.Errorf(format, v...))
			// log.Fatalf(format, v...)
		}
	}
	sq, err := sqinn.NewSqinn(sqinnPath, sqinn.StdLogger{})
	assert(err == nil, "%s", err)
	// open db
	err = sq.Open(dbFile)
	assert(err == nil, "%s", err)
	err = sq.Exec("DROP TABLE IF EXISTS users", 1, 0, nil)
	assert(err == nil, "%s", err)
	err = sq.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, image BLOB)", 1, 0, nil)
	assert(err == nil, "%s", err)
	// insert
	id := 1
	image := make([]byte, 64)
	for i := 0; i < len(image); i++ {
		image[i] = byte(i)
	}
	values := []interface{}{id, image}
	err = sq.Exec("INSERT INTO users (id,image) VALUES(?,?)", 1, 2, values)
	assert(err == nil, "%s", err)
	// query
	sql := "SELECT id, image FROM users ORDER BY id"
	rows, err := sq.Query(sql, nil, []byte{sqinn.VAL_INT, sqinn.VAL_BLOB})
	assert(err == nil, "%s", err)
	assert(len(rows) == 1, "wrong rows %d", len(rows))
	// close and terminate
	err = sq.Close()
	assert(err == nil, "%s", err)
	err = sq.Terminate()
	assert(err == nil, "%s", err)
	log.Printf("TEST %s OK", funcname)
}

func main() {
	// log.SetOutput(ioutil.Discard)
	log.SetFlags(log.Lmicroseconds)
	sqinnPath := "sqinn"
	dbFile := ":memory:"
	flag.StringVar(&sqinnPath, "sqinn", sqinnPath, "name of sqinn executable")
	flag.StringVar(&dbFile, "db", dbFile, "path to db file")
	flag.Parse()
	for _, arg := range flag.Args() {
		if arg == "test" {
			testFunctions(sqinnPath, dbFile, 2)
			testUsers(sqinnPath, dbFile, 2, true)
			testComplex(sqinnPath, dbFile, 2, 2, 2)
			testBlob(sqinnPath, dbFile)
			return
		} else if arg == "bench" {
			testFunctions(sqinnPath, dbFile, 10*1000)
			testUsers(sqinnPath, dbFile, 1000*1000, true)
			testComplex(sqinnPath, dbFile, 100, 100, 10)
			return
		}
	}
	fmt.Printf("no command, want 'test' or 'bench'\n")
}
