package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	// please remove comment to enable go-sqlite3 benchmark
	// _ "github.com/mattn/go-sqlite3"
)

func testUsers(dbFile string, nusers int, bindRating bool) {
	funcname := "testUsers"
	log.Printf("TEST %s", funcname)
	log.Printf("dbFile=%s, nusers=%d, bindRating=%t", dbFile, nusers, bindRating)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// open db
	db, err := sql.Open("sqlite3", dbFile)
	check(err)
	defer db.Close()
	// prepare schema
	_, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY NOT NULL, name VARCHAR, age INTEGER, rating REAL)")
	check(err)
	// insert users
	t1 := time.Now()
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare("INSERT INTO users (id, name, age, rating) VALUES (?,?,?,?)")
	check(err)
	for i := 0; i < nusers; i++ {
		id := i + 1
		name := fmt.Sprintf("User_%d", id)
		age := 33 + i
		if bindRating {
			rating := 0.13 * float64(i+1)
			_, err = stmt.Exec(id, name, age, rating)
			check(err)
		} else {
			_, err = stmt.Exec(id, name, age, nil)
			check(err)
		}
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	t2 := time.Now()
	// query users
	rows, err := db.Query("SELECT id, name, age, rating FROM users ORDER BY id")
	check(err)
	nrows := 0
	var id sql.NullInt32
	var name sql.NullString
	var age sql.NullInt32
	var rating sql.NullFloat64
	for rows.Next() {
		nrows++
		err = rows.Scan(&id, &name, &age, &rating)
		check(err)

	}
	log.Printf("fetched %d rows", nrows)
	t3 := time.Now()
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func testComplex(dbFile string, nprofiles, nusers, nlocations int) {
	funcname := "testComplex"
	log.Printf("TEST %s", funcname)
	log.Printf("dbFile=%s, nprofiles, nusers, nlocations = %d, %d, %d", dbFile, nprofiles, nusers, nlocations)
	// make sure db doesn't exist
	os.Remove(dbFile)
	// open db
	db, err := sql.Open("sqlite3", dbFile)
	check(err)
	defer db.Close()
	// prepare schema
	_, err = db.Exec("PRAGMA foreign_keys=1")
	check(err)
	_, err = db.Exec("CREATE TABLE profiles (id VARCHAR PRIMARY KEY NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL)")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_profiles_name ON profiles(name);")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_profiles_active ON profiles(active);")
	check(err)
	_, err = db.Exec("CREATE TABLE users (id VARCHAR PRIMARY KEY NOT NULL, profileId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (profileId) REFERENCES profiles(id))")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_users_profileId ON users(profileId);")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_users_name ON users(name);")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_users_active ON users(active);")
	check(err)
	_, err = db.Exec("CREATE TABLE locations (id VARCHAR PRIMARY KEY NOT NULL, userId VARCHAR NOT NULL, name VARCHAR NOT NULL, active BOOL NOT NULL, FOREIGN KEY (userId) REFERENCES users(id))")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_locations_userId ON locations(userId);")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_locations_name ON locations(name);")
	check(err)
	_, err = db.Exec("CREATE INDEX idx_locations_active ON locations(active);")
	check(err)
	// insert
	t1 := time.Now()
	tx, err := db.Begin()
	check(err)
	stmt, err := tx.Prepare("INSERT INTO profiles (id,name,active) VALUES(?,?,?)")
	for p := 0; p < nprofiles; p++ {
		profileId := fmt.Sprintf("profile_%d", p)
		name := fmt.Sprintf("ProfileGo %d", p)
		active := p % 2
		_, err = stmt.Exec(profileId, name, active)
		check(err)
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	//
	tx, err = db.Begin()
	check(err)
	stmt, err = tx.Prepare("INSERT INTO users (id,profileId,name,active) VALUES(?,?,?,?)")
	for p := 0; p < nprofiles; p++ {
		profileId := fmt.Sprintf("profile_%d", p)
		for u := 0; u < nusers; u++ {
			userId := fmt.Sprintf("user_%d_%d", p, u)
			name := fmt.Sprintf("User %d %d", p, u)
			active := u % 2
			_, err = stmt.Exec(userId, profileId, name, active)
			check(err)
		}
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	//
	tx, err = db.Begin()
	check(err)
	stmt, err = tx.Prepare("INSERT INTO locations (id,userId,name,active) VALUES(?,?,?,?)")
	for p := 0; p < nprofiles; p++ {
		for u := 0; u < nusers; u++ {
			userId := fmt.Sprintf("user_%d_%d", p, u)
			for l := 0; l < nlocations; l++ {
				locationId := fmt.Sprintf("location_%d_%d_%d", p, u, l)
				name := fmt.Sprintf("Location %d %d %d", p, u, l)
				active := l % 2
				_, err = stmt.Exec(locationId, userId, name, active)
				check(err)
			}
		}
	}
	err = stmt.Close()
	check(err)
	err = tx.Commit()
	check(err)
	t2 := time.Now()
	// query
	query := "SELECT locations.id, locations.userId, locations.name, locations.active, users.id, users.profileId, users.name, users.active, profiles.id, profiles.name, profiles.active " +
		"FROM locations " +
		"LEFT JOIN users ON users.id = locations.userId " +
		"LEFT JOIN profiles ON profiles.id = users.profileId " +
		"WHERE locations.active = ? OR locations.active = ? " +
		"ORDER BY locations.name, locations.id, users.name, users.id, profiles.name, profiles.id"
	rows, err := db.Query(query, 0, 1)
	check(err)
	nrows := 0
	var locationId sql.NullInt32
	var locationUserId sql.NullInt32
	var locationName sql.NullString
	var locationActive sql.NullBool
	var userId sql.NullInt32
	var userProfileId sql.NullInt32
	var userName sql.NullString
	var userActive sql.NullBool
	var profileId sql.NullInt32
	var profileName sql.NullString
	var profileActive sql.NullBool
	for rows.Next() {
		nrows++
		rows.Scan(
			&locationId,
			&locationUserId,
			&locationName,
			&locationActive,
			&userId,
			&userProfileId,
			&userName,
			&userActive,
			&profileId,
			&profileName,
			&profileActive,
		)
	}
	log.Printf("fetched %d rows", nrows)
	t3 := time.Now()
	// done
	log.Printf("insert took %s", t2.Sub(t1))
	log.Printf("query took %s", t3.Sub(t2))
	log.Printf("TEST %s OK", funcname)
}

func main() {
	dbFile := ":memory:"
	flag.StringVar(&dbFile, "db", dbFile, "path to db file")
	flag.Parse()
	testUsers(dbFile, 1000*1000, false)
	testUsers(dbFile, 1000*1000, true)
	testComplex(dbFile, 100, 100, 10)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
