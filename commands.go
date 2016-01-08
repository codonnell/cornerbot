package main

import (
	// "fmt"
	"io/ioutil"
	"strings"
)

type ActionCommand struct {
	Command string
	Message string
}

type MessageCommand struct {
	Command string
	Message string
}

func ReadActions() []ActionCommand {
	b, e := ioutil.ReadFile("actions.txt")
	if e != nil {
		return []ActionCommand{}
	}
	s := string(b)
	actionLines := strings.Split(s, "\n")
	var actions []ActionCommand
	for _, l := range actionLines {
		if len(l) < 2 {
			continue
		}
		sep := l[0]
		splitAction := strings.Split(l[1:], string(sep))
		actions = append(actions, ActionCommand{Command: splitAction[0], Message: splitAction[1]})
	}
	return actions
}

func ReadMessages() []MessageCommand {
	b, e := ioutil.ReadFile("messages.txt")
	if e != nil {
		return []MessageCommand{}
	}
	s := string(b)
	actionLines := strings.Split(s, "\n")
	var messages []MessageCommand
	for _, l := range actionLines {
		if len(l) < 2 {
			continue
		}
		sep := l[0]
		splitMessage := strings.Split(l[1:], string(sep))
		messages = append(messages, MessageCommand{Command: splitMessage[0], Message: splitMessage[1]})
	}
	return messages
}

// func main() {
// 	actions := ReadActions()
// 	for _, a := range actions {
// 		fmt.Println("command: " + a.Command + ", message: " + a.Message)
// 	}
// }
