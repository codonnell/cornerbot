package main

//
// import (
// 	"database/sql"
// 	"fmt"
// 	_ "github.com/lib/pq"
// 	"log"
// )
//
// type CornerDB struct {
// 	db *sql.DB
// }
//
// func Connect() *sql.DB {
// 	db, err := sql.Open("postgres",
// 		"dbname=chris user=chris host=/var/run/postgresql sslmode=disable")
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil
// 	}
// 	defer db.Close()
// 	return db
// }
//
// func (db *CornerDB) AddUser(nick string) {
// 	rows, err := db.Query("select nick from users where nick = $1", nick)
// 	if err != nil {
// 		log.Fatal(err)
// 		return nil
// 	}
// 	defer rows.Close()
// 	var dbNick string
// 	rows.Next()
// 	err := rows.Scan(&dbNick)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println(dbNick)
// }
//
// func main() {
// 	db, err := sql.Open("postgres",
// 		"dbname=chris user=chris host=/var/run/postgresql sslmode=disable")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer db.Close()
// 	err = db.Ping()
// 	if err != nil {
// 		fmt.Println("Connection error: %s\n", err)
// 		return
// 	}
//
// 	rows, err := db.Query("SELECT * FROM users")
// 	defer rows.Close()
// 	for rows.Next() {
// 		var nick string
// 		var points int
// 		err = rows.Scan(&nick, &points)
// 		if err != nil {
// 			fmt.Printf("Connection error: %s\n", err)
// 			return
// 		}
// 		fmt.Printf("Nick: %s\tPoints: %d\n", nick, points)
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		fmt.Printf("Connection error: %s\n", err)
// 		return
// 	}
//
// }
