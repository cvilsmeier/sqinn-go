/*
A benchmark for sqinn-go.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/cvilsmeier/sqinn-go/sqinn"
)

func benchUsers(sqinnPath, dbFile string, nusers int, bindRating bool) {
	funcname := "benchUsers"
	log.Printf("BENCH %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d, bindRating=%t", sqinnPath, dbFile, nusers, bindRating)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	err = sq.Open(dbFile)
	check(err)
	// prepare schema
	_, err = sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)")
	// insert users
	t1 := time.Now()
	_, err = sq.Exec("BEGIN TRANSACTION", 1, 0, nil)
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
	_, err = sq.Exec("INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)", nusers, 4, values)
	_, err = sq.ExecOne("COMMIT")
	t2 := time.Now()
	// query users
	colTypes := []byte{sqinn.ValInt, sqinn.ValText, sqinn.ValInt, sqinn.ValDouble}
	rows, err := sq.Query("SELECT id, name, age, rating FROM users ORDER BY id", nil, colTypes)
	check(err)
	if len(rows) != nusers {
		log.Printf("want %v rows but was %v", nusers, len(rows))
	}
	t3 := time.Now()
	// close db
	err = sq.Close()
	check(err)
	// terminate sqinn
	err = sq.Terminate()
	check(err)
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("BENCH %s OK", funcname)
}

func benchComplexSchema(sqinnPath, dbFile string, nprofiles, nusers, nlocations int) {
	funcname := "benchComplexSchema"
	log.Printf("BENCH %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nprofiles, nusers, nlocations = %d, %d, %d", sqinnPath, dbFile, nprofiles, nusers, nlocations)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	// open db
	check(sq.Open(dbFile))
	_, err = sq.ExecOne("PRAGMA foreign_keys=1")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS locations")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS users")
	check(err)
	_, err = sq.ExecOne("DROP TABLE IF EXISTS profiles")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE profiles (id VARCHAR PRIMARY KEY NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL)")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_profiles_name ON profiles(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_profiles_active ON profiles(active);")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE users (id VARCHAR PRIMARY KEY NOT NULL, profileId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (profileId) REFERENCES profiles(id))")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_profileId ON users(profileId);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_name ON users(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_users_active ON users(active);")
	check(err)
	_, err = sq.ExecOne("CREATE TABLE locations (id VARCHAR PRIMARY KEY NOT NULL, userId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (userId) REFERENCES users(id))")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_userId ON locations(userId);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_name ON locations(name);")
	check(err)
	_, err = sq.ExecOne("CREATE INDEX idx_locations_active ON locations(active);")
	check(err)
	// insert
	t1 := time.Now()
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values := make([]interface{}, 0, nprofiles*3)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("ProfileGo %d", p)
		active := p % 2
		values = append(values, profileID, name, active)
	}
	_, err = sq.Exec("INSERT INTO profiles (id,name,active) VALUES(?,?,?)", nprofiles, 3, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values = make([]interface{}, 0, nprofiles*nusers*4)
	for p := 0; p < nprofiles; p++ {
		profileID := fmt.Sprintf("profile_%d", p)
		for u := 0; u < nusers; u++ {
			userID := fmt.Sprintf("user_%d_%d", p, u)
			name := fmt.Sprintf("User %d %d", p, u)
			active := u % 2
			values = append(values, userID, profileID, name, active)
		}
	}
	_, err = sq.Exec("INSERT INTO users (id,profileId,name,active) VALUES(?,?,?,?)", nprofiles*nusers, 4, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	_, err = sq.ExecOne("BEGIN TRANSACTION")
	check(err)
	values = make([]interface{}, 0, nprofiles*nusers*nlocations*4)
	for p := 0; p < nprofiles; p++ {
		for u := 0; u < nusers; u++ {
			userID := fmt.Sprintf("user_%d_%d", p, u)
			for l := 0; l < nlocations; l++ {
				locationID := fmt.Sprintf("location_%d_%d_%d", p, u, l)
				name := fmt.Sprintf("Location %d %d %d", p, u, l)
				active := l % 2
				values = append(values, locationID, userID, name, active)
			}
		}
	}
	_, err = sq.Exec("INSERT INTO locations (id,userId,name,active) VALUES(?,?,?,?)", nprofiles*nusers*nlocations, 4, values)
	check(err)
	_, err = sq.Exec("COMMIT", 1, 0, nil)
	check(err)
	t2 := time.Now()
	// query
	sql := "SELECT locations.id, locations.userId, locations.name, locations.active, users.id, users.profileId, users.name, users.active, profiles.id, profiles.name, profiles.active " +
		"FROM locations " +
		"LEFT JOIN users ON users.id = locations.userId " +
		"LEFT JOIN profiles ON profiles.id = users.profileId " +
		"WHERE locations.active = ? OR locations.active = ? " +
		"ORDER BY locations.name, locations.id, users.name, users.id, profiles.name, profiles.id"
	rows, err := sq.Query(sql, []interface{}{0, 1}, []byte{sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValText, sqinn.ValInt, sqinn.ValText, sqinn.ValText, sqinn.ValInt})
	check(err)
	expectedRows := nprofiles * nusers * nlocations
	if len(rows) != expectedRows {
		log.Fatalf("expected %v rows but was %v", expectedRows, len(rows))
	}
	t3 := time.Now()
	// close and terminate
	check(sq.Close())
	check(sq.Terminate())
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("BENCH %s OK", funcname)
}

func benchConcurrent(sqinnPath, dbFile string, nusers, nworkers int) {
	funcname := "benchConcurrent"
	log.Printf("BENCH %s", funcname)
	log.Printf("sqinnPath=%s, dbFile=%s, nusers=%d, nworkers=%d", sqinnPath, dbFile, nusers, nworkers)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// launch sqinn
	sq, err := sqinn.Launch(sqinn.Options{
		SqinnPath: sqinnPath,
	})
	check(err)
	defer sq.Terminate()
	// open db
	check(sq.Open(dbFile))
	defer sq.Close()
	// prepare schema
	_, err = sq.ExecOne("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR)")
	check(err)
	// insert nusers
	_, err = sq.ExecOne("BEGIN")
	check(err)
	values := make([]interface{}, 0, nusers*2)
	for u := 0; u < nusers; u++ {
		id := u + 1
		name := fmt.Sprintf("User %d", u)
		values = append(values, id, name)
	}
	_, err = sq.Exec("INSERT INTO users (id,name) VALUES(?,?)", nusers, 2, values)
	check(err)
	_, err = sq.ExecOne("COMMIT")
	check(err)
	// query
	t1 := time.Now()
	var wg sync.WaitGroup
	wg.Add(nworkers)
	for w := 0; w < nworkers; w++ {
		go func(w int) {
			defer wg.Done()
			// log.Printf("worker %v start", w)
			// defer log.Printf("worker %v end", w)
			// launch sqinn
			sq, err := sqinn.Launch(sqinn.Options{
				SqinnPath: sqinnPath,
			})
			check(err)
			defer sq.Terminate()
			// open db
			err = sq.Open(dbFile)
			check(err)
			defer sq.Close()
			// set busy timeout
			_, err = sq.ExecOne("PRAGMA busy_timeout = 10000;")
			check(err)
			// query
			rows, err := sq.Query("SELECT id, name FROM users ORDER BY id", nil, []byte{sqinn.ValInt, sqinn.ValText})
			if err != nil {
				log.Fatalf("worker %v: %v", w, err)
			}
			nrows := len(rows)
			// log.Printf("worker %v has %v rows", w, nrows)
			if nrows != nusers {
				log.Fatalf("worker %v: want %v rows but was %v", w, nusers, nrows)
			}
		}(w)
	}
	wg.Wait()
	t2 := time.Now()
	// done
	log.Printf("queries took %s", t2.Sub(t1))
	log.Printf("BENCH %s DONE", funcname)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	sqinnPath := os.Getenv("SQINN_PATH")
	dbFile := ""
	flag.StringVar(&sqinnPath, "sqinn", sqinnPath, "path to sqinn executable")
	flag.StringVar(&dbFile, "db", dbFile, "path to db file")
	flag.Parse()
	if dbFile == "" {
		log.Fatalf("no dbFile, please set -db flag")
	}
	benchUsers(sqinnPath, dbFile, 1000*1000, false)
	benchUsers(sqinnPath, dbFile, 1000*1000, true)
	benchComplexSchema(sqinnPath, dbFile, 200, 100, 10)
	benchConcurrent(sqinnPath, dbFile, 1000*1000, 2)
	benchConcurrent(sqinnPath, dbFile, 1000*1000, 4)
	benchConcurrent(sqinnPath, dbFile, 1000*1000, 8)
}
