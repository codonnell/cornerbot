package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Command struct {
	Name    string
	Message string
}

func readLines(file string) ([]string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	s := string(b)
	lines := strings.Split(s, "\n")
	blankLines := []int{}
	for i, line := range lines {
		if line == "" {
			blankLines = append(blankLines, i)
		}
	}
	for _, i := range blankLines {
		lines = append(lines[:i], lines[i+1:]...)
	}
	return lines, nil
}

func AddCommandToFile(command Command, commandFile string) {
	seps := []string{"#", "/", "^", "$", "@", "*", "~"}
	var sep string
	for _, s := range seps {
		if !strings.Contains(command.Message, s) && !strings.Contains(command.Name, s) {
			sep = s
			fmt.Println("Choosing sep " + sep)
			break
		}
	}
	f, err := os.OpenFile(commandFile, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
	}
	defer f.Close()
	_, err = f.WriteString("\n" + sep + command.Name + sep + command.Message)
	if err != nil {
		fmt.Println("Error writing to file: ", err.Error())
	}
}

func DeleteCommandFromFile(name string, commandFile string) {
	lines, err := readLines(commandFile)
	fmt.Println("Before removal:")
	fmt.Println(lines)
	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
	}
	toRemove := -1
	for i, line := range lines {
		sep := string(line[0])
		if strings.HasPrefix(line, sep+name+sep) {
			toRemove = i
		}
	}
	if toRemove != -1 {
		lines = append(lines[:toRemove], lines[toRemove+1:]...)
	}
	fmt.Println("After removal:")
	fmt.Println(lines)
	err = ioutil.WriteFile("actions.txt", []byte(strings.Join(lines, "\n")), 0666)
	if err != nil {
		fmt.Println("Error writing to file: " + err.Error())
	}
}

func ReadCommandsFromFile(commandFile string) []Command {
	lines, err := readLines(commandFile)
	if err != nil {
		fmt.Println("Error reading file: " + err.Error())
	}
	var commands []Command
	for _, l := range lines {
		if len(l) < 2 {
			continue
		}
		sep := l[0]
		splitLine := strings.Split(l[1:], string(sep))
		commands = append(commands, Command{Name: splitLine[0], Message: splitLine[1]})
	}
	return commands
}

// func main() {
// 	actions := ReadActions()
// 	for _, a := range actions {
// 		fmt.Println("command: " + a.Command + ", message: " + a.Message)
// 	}
// }
