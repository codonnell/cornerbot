package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	irc "github.com/fluffle/goirc/client"
	"log"
	"os"
	"strings"
)

type BotConfig struct {
	Host     string   `json:"host"`
	Channels []string `json:"channels"`
	User     string   `json:"user"`
	Nick     string   `json:"nick"`
	Owner    string   `json:"owner"`
}

var DB CornerDB
var config BotConfig

// var host *string = flag.String("host", "irc.utonet.org", "IRC server")
// var channel *string = flag.String("channel", "#thecorner", "IRC channel")
// var reqiRainbowChan *string = flag.String("reqiRainbowChan", "#supersecretprivatechan", "rainbow and reqi chan")

func main() {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	config = BotConfig{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}
	DB = CornerDB{Connect()}
	defer DB.db.Close()

	// create new IRC connection
	c := irc.SimpleClient(config.User, config.Nick)
	c.EnableStateTracking()
	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			for _, c := range config.Channels {
				conn.Join(c)
			}
		})

	// Set up a handler to notify of disconnect events.
	quit := make(chan bool)
	c.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// Add handlers for the bot commands in handlers.go
	addBotHandlers(c)

	// set up a goroutine to read commands from stdin
	in := make(chan string, 4)
	reallyquit := false
	go func() {
		con := bufio.NewReader(os.Stdin)
		for {
			s, err := con.ReadString('\n')
			if err != nil {
				// wha?, maybe ctrl-D...
				close(in)
				break
			}
			// no point in sending empty lines down the channel
			if len(s) > 2 {
				in <- s[0 : len(s)-1]
			}
		}
	}()

	// set up a goroutine to do parsey things with the stuff from stdin
	go func() {
		for cmd := range in {
			if cmd[0] == ':' {
				switch idx := strings.Index(cmd, " "); {
				case cmd[1] == 'd':
					fmt.Printf(c.String())
				case cmd[1] == 'f':
					if len(cmd) > 2 && cmd[2] == 'e' {
						// enable flooding
						c.Config().Flood = true
					} else if len(cmd) > 2 && cmd[2] == 'd' {
						// disable flooding
						c.Config().Flood = false
					}
					for i := 0; i < 20; i++ {
						c.Privmsg("#", "flood test!")
					}
				case idx == -1:
					continue
				case cmd[1] == 'q':
					reallyquit = true
					c.Quit(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'j':
					c.Join(cmd[idx+1 : len(cmd)])
				case cmd[1] == 'p':
					c.Part(cmd[idx+1 : len(cmd)])
				}
			} else {
				c.Raw(cmd)
			}
		}
	}()

	for !reallyquit {
		// connect to server
		if err := c.ConnectTo(config.Host); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}

		// wait on quit channel
		<-quit
	}
}
