package main

import (
	// "fmt"
	"regexp"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

// NOTE: Do not call this function synchronously from a handler; it will freeze
// the bot.
func isIdentified(conn *irc.Conn, nick string) bool {
	authChan := make(chan bool, 1)
	handler := conn.HandleFunc("notice", func(conn *irc.Conn, line *irc.Line) {
		isStatus := regexp.MustCompile("STATUS " + nick + " (\\d)")
		if strings.ToLower(line.Target()) != "nickserv" {
			return
		}
		matches := isStatus.FindStringSubmatch(line.Text())
		if len(matches) < 2 {
			return
		}
		if matches[1] == "3" {
			authChan <- true
		} else {
			authChan <- false
		}
	})
	conn.Privmsg("nickserv", "status "+nick)
	isAuthed := <-authChan
	handler.Remove()
	return isAuthed
}
