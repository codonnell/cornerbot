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

type Command struct {
	Name    string
	Message string
	Type string
}

func Connect() *sql.DB {
	db, err := sql.Open("sqlite3", "./bot.db")
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
		db.CreateUser(nick)
		db.SetPoints(nick, points)
	} else {
		db.SetPoints(nick, user.points+points)
	}
}

func (db *CornerDB) GetCommand(name string) *Command {
	rows, err := db.db.Query("select name,message,type from commands where name = ?", name)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer rows.Close()
	var dbName, dbMessage, dbType string
	if rows.Next() {
		err = rows.Scan(&dbName, &dbMessage, &dbType)
		checkErr(err)
	} else {
		return nil
	}
	return &Command{dbName, dbMessage, dbType}
}

func (db *CornerDB) AddCommand(command Command) bool {
	if db.GetCommand(command.Name) != nil {
		return false
	}
	_, err := db.db.Exec("insert into commands(name, message, type) values(?, ?, ?)", command.Name, command.Message, command.Type)
	return (err == nil)
}

func (db *CornerDB) AllCommands(cmdType string) []Command {
	rows, err := db.db.Query("select name,message from commands where type = ?", cmdType)
	checkErr(err)
	defer rows.Close()
	var name, message string
	commands := []Command{}
	for rows.Next() {
		err = rows.Scan(&name, &message)
		checkErr(err)
		commands = append(commands, Command{name, message, cmdType})
	}
	err = rows.Err()
	checkErr(err)
	return commands
}

func (db *CornerDB) DeleteCommand(name string) bool {
	_, err := db.db.Exec("delete from commands where name = ?", name)
	return (err == nil)
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
