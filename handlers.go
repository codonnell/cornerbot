package main

import (
	irc "github.com/fluffle/goirc/client"
	"regexp"
)

var botHandlers = []irc.HandlerFunc{
	Heart,
	RandomPage,
	Hug,
}

func addBotHandlers(conn *irc.Conn) {
	for _, h := range botHandlers {
		conn.HandleFunc("privmsg", h)
	}
}

func Heart(conn *irc.Conn, line *irc.Line) {
	containsHeart := regexp.MustCompile(`.*<3.*`)
	if containsHeart.MatchString(line.Text()) {
		conn.Privmsg(line.Target(), "<3")
	}
}

func Hug(conn *irc.Conn, line *irc.Line) {
	isHug := regexp.MustCompile(`!hug (.+)`)
	matches := isHug.FindStringSubmatch(line.Text())
	if len(matches) < 2 {
		return
	}
	if matches[1] == "all" {
		conn.Action(line.Target(), "hugs everyone!")
	} else {
		conn.Action(line.Target(), "hugs "+matches[1])
	}
}

func RandomPage(conn *irc.Conn, line *irc.Line) {
	if line.Text() != "!til..." {
		return
	}
	id, err := RandomPageID()
	if err != nil {
		return
	}
	url, err := PageUrl(id)
	if err != nil {
		return
	}
	conn.Privmsg(line.Target(), url)
}
