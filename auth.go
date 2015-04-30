package main

import (
	"regexp"

	irc "github.com/fluffle/goirc/client"
)

func isIdentified(conn *irc.Conn, nick string) bool {
	authChan := make(chan bool)
	remover := conn.HandleFunc("privmsg", func(conn *irc.Conn, line *irc.Line) {
		isStatus := regexp.MustCompile("status " + nick + " \\d")
		if line.Target() != "NickServ" {
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
	conn.Privmsg("NickServ", "status "+nick)
	isAuthed := <-authChan
	remover.Remove()
	return isAuthed
}
