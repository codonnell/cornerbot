package main

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

type KeyedHandler struct {
	Key     string
	Handler irc.HandlerFunc
}

var botHandlers = []irc.HandlerFunc{
	AddPointsHandler,
	GetPointsHandler,
	Printer,
	Heart,
	RandomPage,
	Hug,
	SelfEsteem,
	Glitter,
	Slap,
	Lurve,
	Identified,
	Cookie,
	AddCommandHandler,
	DeleteCommandHandler,
	Say,
	ListChannels,
	CreateMessage("rainbowsaurus", Colorize("rRra@Aa.wwWWw.rrRr")),
}

var keyedHandlers = make(map[string]irc.Remover)

// NOTE: Do not call this synchronously from a handler. It will freeze the bot.
func AddKeyedHandler(conn *irc.Conn, key string, handler irc.HandlerFunc) {
	keyedHandlers[key] = conn.HandleFunc("privmsg", handler)
}

func addBotHandlers(conn *irc.Conn) {
	for _, a := range DB.AllCommands("action") {
		AddKeyedHandler(conn, a.Name, CreateAction(a.Name, a.Message))
	}
	for _, m := range DB.AllCommands("message") {
		AddKeyedHandler(conn, m.Name, CreateMessage(m.Name, m.Message))
	}
	for _, h := range botHandlers {
		conn.HandleFunc("privmsg", h)
	}
	conn.HandleFunc("notice", Printer)
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
			if strings.Contains(message, "%s") || strings.Contains(message, "%[1]s") {
				conn.Action(line.Target(), fmt.Sprintf(message, matches[1]))
			} else {
				conn.Action(line.Target(), message)
			}
		}
	}
}

func CreateMessage(name string, message string) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		isMessage := regexp.MustCompile("!" + name + " ?(\\S+)?")
		matches := isMessage.FindStringSubmatch(line.Text())
		if len(matches) > 0 {
			fmt.Println("name: " + name)
			fmt.Println("message: " + message)
			if strings.Contains(message, "%s") {
				conn.Privmsg(line.Target(), fmt.Sprintf(message, line.Nick))
			} else {
				conn.Privmsg(line.Target(), message)
			}
		}
	}
}

func AddCommandHandler(conn *irc.Conn, line *irc.Line) {
	isAddAction := regexp.MustCompile(`!addaction (\S+) (.+)`)
	isAddMessage := regexp.MustCompile(`!addmessage (\S+) (.+)`)
	actionMatches := isAddAction.FindStringSubmatch(line.Text())
	messageMatches := isAddMessage.FindStringSubmatch(line.Text())
	if len(actionMatches) == 0 && len(messageMatches) == 0 {
		return
	}
	var cmdType, name, message string
	var handler irc.HandlerFunc
	if len(actionMatches) > 0 {
		cmdType = "action"
		name = actionMatches[1]
		message = actionMatches[2]
		handler = CreateAction(name, message)
	} else {
		cmdType = "message"
		name = messageMatches[1]
		message = messageMatches[2]
		handler = CreateMessage(name, message)
	}
	go func() {
		if line.Nick == config.Owner && isIdentified(conn, config.Owner) {
			command := Command{name, message, cmdType}
			if _, exists := keyedHandlers[command.Name]; exists {
				conn.Privmsg(line.Target(), "The "+command.Name+" command already exists. You must delete it with !delaction "+command.Name)
				return
			}
			fmt.Println("Adding action with command " + command.Name + " and message " + command.Message)
			conn.Privmsg(line.Target(), "The "+command.Name+" action has been added")
			DB.AddCommand(Command{name, message, cmdType})
			AddKeyedHandler(conn, command.Name, handler)
		} else {
			conn.Privmsg(line.Target(), "Only the bot owner can add commands.")
		}
	}()
}

func DeleteCommandHandler(conn *irc.Conn, line *irc.Line) {
	isDelAction := regexp.MustCompile(`!delaction (\S+)`)
	isDelMessage := regexp.MustCompile(`!delmessage (\S+)`)
	actionMatches := isDelAction.FindStringSubmatch(line.Text())
	messageMatches := isDelMessage.FindStringSubmatch(line.Text())
	if len(actionMatches) == 0 && len(messageMatches) == 0 {
		return
	}
	var command string
	if len(actionMatches) > 0 {
		command = actionMatches[1]
	} else {
		command = messageMatches[1]
	}
	go func() {
		if line.Nick == config.Owner && isIdentified(conn, config.Owner) {
			if remover, exists := keyedHandlers[command]; exists {
				DB.DeleteCommand(command)
				remover.Remove()
				delete(keyedHandlers, command)
				conn.Privmsg(line.Target(), "The "+command+" command has been deleted")
			} else {
				conn.Privmsg(line.Target(), "The "+command+" command does not exist or cannot be deleted.")
			}
		} else {
			conn.Privmsg(line.Target(), "Only the bot owner can delete commands.")
		}
	}()
}

func Printer(conn *irc.Conn, line *irc.Line) {
	fmt.Println(line.Target(), ": ", line.Text())
}

func Identified(conn *irc.Conn, line *irc.Line) {
	isIdentify := regexp.MustCompile(`!identified (\S+)`)
	matches := isIdentify.FindStringSubmatch(line.Text())
	if len(matches) == 2 {
		fmt.Println("Calls isIdentified.")
		go IdentifyMsg(conn, line.Target(), matches[1])
	}
}

func IdentifyMsg(conn *irc.Conn, target string, nick string) {
	identified := isIdentified(conn, nick)
	if identified {
		conn.Privmsg(target, nick+" is identified.")
	} else {
		conn.Privmsg(target, nick+" is not identified.")
	}
}

func ListChannels(conn *irc.Conn, line *irc.Line) {
	if line.Text() != "!listchannels" {
		return
	}
	go func() {
		if line.Nick == config.Owner && isIdentified(conn, config.Owner) {
			var channels []string
			for i, c := range config.Channels {
				channels = append(channels, fmt.Sprintf("%d: %s", i, c))
			}
			conn.Privmsg(line.Target(), strings.Join(channels, ", "))
		}
	}()
}

func Say(conn *irc.Conn, line *irc.Line) {
	isSay := regexp.MustCompile(`!say (\d+) (.+)`)
	matches := isSay.FindStringSubmatch(line.Text())
	if len(matches) == 0 {
		return
	}
	go func() {
		if line.Nick == config.Owner && isIdentified(conn, config.Owner) {
			idx, err := strconv.Atoi(matches[1])
			if err != nil {
				conn.Privmsg(line.Target(), "Error parsing channel number")
				return
			}
			if idx >= 0 && idx < len(config.Channels) {
				ircChan := config.Channels[idx]
				conn.Privmsg(ircChan, matches[2])
			} else {
				conn.Privmsg(line.Target(), "I'm not in a channel with that number.")
			}
		}
	}()
}

func Cookie(conn *irc.Conn, line *irc.Line) {
	isCookie := regexp.MustCompile(`!cookie(\s*)(\S*)`)
	matches := isCookie.FindStringSubmatch(line.Text())
	if len(matches) == 0 {
		return
	}
	if matches[2] == "" {
		conn.Action(line.Target(), "gives you a cookie. but then steals it and eats it!")
	} else {
		conn.Action(line.Target(), fmt.Sprintf("gives %s a cookie.", matches[2]))
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

func AddPointsHandler(conn *irc.Conn, line *irc.Line) {
	isAddPoints := regexp.MustCompile(`!addpts (\S+) ([-]?\d+)`)
	matches := isAddPoints.FindStringSubmatch(line.Text())
	if len(matches) != 3 {
		return
	}
	if line.Nick != config.Owner {
		conn.Privmsg(line.Target(), "Hah, nice try imposter!")
		return
	}
	points, err := strconv.ParseInt(matches[2], 10, 64)
	if err != nil {
		fmt.Println("Failed at parsing int")
		return
	}
	go AddPointsHelper(conn, line, matches[1], int(points))
}

func AddPointsHelper(conn *irc.Conn, line *irc.Line, nick string, points int) {
	if isIdentified(conn, config.Owner) {
		if nick == config.Owner {
			conn.Privmsg(line.Target(), nick+" already has more rainbow points than there are numbers in the universe! Trying to give them more points is futile.")
		} else {
			DB.AddPoints(nick, points)
			conn.Privmsg(line.Target(), nick+" has "+strconv.Itoa(DB.GetPoints(nick))+Colorize(" rainbow points."))
		}
	} else {
		conn.Privmsg(line.Target(), "Nice try... I'm onto you.")
	}
}

func GetPointsHandler(conn *irc.Conn, line *irc.Line) {
	isGetPoints := regexp.MustCompile(`!pts (\S+)`)
	matches := isGetPoints.FindStringSubmatch(line.Text())
	if len(matches) != 2 {
		return
	}
	if matches[1] == config.Owner {
		conn.Privmsg(line.Target(), "You wish you had as many "+Colorize("rainbow points")+" as "+config.Owner)
	} else {
		conn.Privmsg(line.Target(), matches[1]+" has "+strconv.Itoa(DB.GetPoints(matches[1]))+Colorize(" rainbow points."))
	}
}
