package main

import (
	"database/sql"
	// "fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type CornerDB struct {
	db *sql.DB
}

type User struct {
	nick   string
	points int
}

func Connect() *sql.DB {
	db, err := sql.Open("sqlite3", "./users.db")
	checkErr(err)
	return db
}

func (db *CornerDB) GetUser(nick string) *User {
	rows, err := db.db.Query("select nick,points from users where nick = $1", nick)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer rows.Close()
	var dbNick string
	var dbPoints int
	if rows.Next() {
		err = rows.Scan(&dbNick, &dbPoints)
		checkErr(err)
	} else {
		return nil
	}
	return &User{dbNick, dbPoints}
}

func (db *CornerDB) GetPoints(nick string) int {
	user := db.GetUser(nick)
	if user == nil {
		return 0
	}
	return user.points
}

func (db *CornerDB) CreateUser(nick string) {
	if db.GetUser(nick) != nil {
		return
	}
	_, err := db.db.Exec("insert into users(nick, points) values(?, ?)", nick, 0)
	checkErr(err)
}

func (db *CornerDB) SetPoints(nick string, points int) {
	user := db.GetUser(nick)
	if user == nil {
		return
	}
	_, err := db.db.Exec("update users set points=? where nick=?", points, nick)
	checkErr(err)
}

func (db *CornerDB) AddPoints(nick string, points int) {
	user := db.GetUser(nick)
	if user == nil {
		return
	}
	db.SetPoints(nick, user.points+points)
}

// func main() {
// 	db := CornerDB{Connect()}
// 	defer db.db.Close()
//
// 	db.CreateUser("testuser")
// 	db.SetPoints("testuser", 100)
// 	db.AddPoints("testuser", -5)
// 	user := db.GetUser("testuser")
// 	if user == nil {
// 		fmt.Println("User does not exist")
// 	} else {
// 		fmt.Printf("Got user: %s, %d\n", user.nick, user.points)
// 	}
// }

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
