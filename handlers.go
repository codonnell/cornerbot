package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	irc "github.com/fluffle/goirc/client"
)

var botHandlers = []irc.HandlerFunc{
	Identified,
	Heart,
	RandomPage,
	Hug,
	SelfEsteem,
	Glitter,
	Slap,
	Lurve,
	Identified,
	CreateAction("tickle", "ties %s up to the bedpost tightly and takes out a feather. Time for some tickles!"),
	CreateAction("lick", "jumps on top of %s and gives them a big slobbery lick!"),
	CreateAction("peck", "sneaks up on %s and delicately pecks them on the cheek"),
	CreateAction("paint", "pulls out the paint set as %s lays naked on the couch posing"),
	CreateAction("landlust", "lusts %s!"),
	CreateAction("beer", "goes to the fridge, grabs a fresh bottle, pops the cap off and hands it to %s. Enjoy!"),
	CreateAction("shuggle", "shuggles %s!"),
	CreateAction("pinch", "sneaks behind %s and gives a little pinch on the butt!"),
	CreateAction("snuggle", "curls up next to %s and snuggles closely"),
	CreateAction("pillowfight", "waits until they are sleeping and SMOTHERS %s WITH A PILLOW!"),
	CreateAction("encourage", "cheers %[1]s on. GOOOOOO %[1]s! YOU GOT DIS!"),
	CreateAction("bearhug", "opens the bear cage and watches as %s gets mauled! AWW! What a big hug!"),
	CreateAction("pat", "pats %s on head, good jooob!"),
	CreateAction("defenestrate", "tosses %s out the window!"),
}

func addBotHandlers(conn *irc.Conn) {
	for _, h := range botHandlers {
		conn.HandleFunc("privmsg", h)
	}
}

func Colorize(text string) string {
	var buffer bytes.Buffer

	for i, char := range text {
		buffer.WriteString("\x03")
		buffer.WriteString(strconv.Itoa((i % 14) + 2))
		buffer.WriteRune(char)
	}
	buffer.WriteString("\x03")
	return buffer.String()
}

func CreateAction(name string, message string) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		isAction := regexp.MustCompile("!" + name + " (\\S+)")
		matches := isAction.FindStringSubmatch(line.Text())
		if len(matches) < 2 {
			return
		} else {
			conn.Action(line.Target(), fmt.Sprintf(message, matches[1]))
		}
	}
}

func Identified(conn *irc.Conn, line *irc.Line) {
	isIdentify := regexp.MustCompile(`!identified (\S+)`)
	matches := isIdentify.FindStringSubmatch(line.Text())
	if len(matches) == 2 {
		identified := isIdentified(conn, matches[1])
		if identified {
			conn.Privmsg(line.Target(), matches[1]+" is identified.")
		} else {
			conn.Privmsg(line.Target(), matches[1]+" is not identified.")
		}
	}
}

func Lurve(conn *irc.Conn, line *irc.Line) {
	isLurve := regexp.MustCompile(`!lurve (\S+)`)
	matches := isLurve.FindStringSubmatch(line.Text())
	if len(matches) == 2 {
		var buffer bytes.Buffer
		for i := 0; i < 10; i++ {
			buffer.WriteString("\x03")
			if i%2 == 0 {
				buffer.WriteString("4")
			} else {
				buffer.WriteString("13")
			}
			buffer.WriteString("❤")
		}
		buffer.WriteString("\x03")
		hearts := buffer.String()
		conn.Privmsg(line.Target(), hearts+matches[1]+hearts)
	}
}

func Glitter(conn *irc.Conn, line *irc.Line) {
	containsGlitter := regexp.MustCompile(`!glitter`)
	if containsGlitter.MatchString(line.Text()) {
		conn.Action(line.Target(), "throws some "+Colorize("¸¸.•*¨*•¸¸.•*¨*• rainbow glitter ¸¸.•*¨*•¸¸.•*¨*•")+" into the air!")
	}
}

func Heart(conn *irc.Conn, line *irc.Line) {
	containsHeart := regexp.MustCompile(`.*<3.*`)
	if containsHeart.MatchString(line.Text()) {
		conn.Privmsg(line.Target(), "<3")
	}
}

func Hug(conn *irc.Conn, line *irc.Line) {
	isHug := regexp.MustCompile(`!hug (\S+)`)
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

func Slap(conn *irc.Conn, line *irc.Line) {
	isSlap := regexp.MustCompile(`!slap (\S)+`)
	matches := isSlap.FindStringSubmatch(line.Text())
	if len(matches) > 0 {
		conn.Action(line.Target(), "bops "+line.Nick+" on the head and reminds them that violence is bad!")
	}
}

func SelfEsteem(conn *irc.Conn, line *irc.Line) {
	isEsteem := regexp.MustCompile(`!selfesteem`)
	matches := isEsteem.FindStringSubmatch(line.Text())
	if len(matches) > 0 {
		conn.Privmsg(line.Target(), "GOOOOOO "+line.Nick+"! YOU GOT DIS!")
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
