package main

import "math/rand"
import "sort"
import "bytes"
import "time"
import "regexp"
import "fmt"
import irc "github.com/fluffle/goirc/client"

const (
	End = iota
	Start
	Join
	Leave
	Error
)

func ProcessLottoMessages(conn *irc.Conn, channel string, messages chan(LottoMessage), close chan(int)) {
	var ticker = time.NewTicker(2 * time.Second)
	messageKeys := make([]int, 0)
	messageStore := make(map[int][]LottoMessage)
	for {
		select {
		case <- close:
			fmt.Println("Processing close message")
			for k := range messageStore {
				conn.Privmsg(channel, formatLottoMessage(k, messageStore[k]))
			}
			return
		case <-ticker.C:
			if len(messageKeys) > 0 {
				sort.Ints(messageKeys)
				fmt.Println(messageKeys)
				fmt.Println(messageStore)
				messageType := messageKeys[0]
				toSend := messageStore[messageType]
				conn.Privmsg(channel, formatLottoMessage(messageType, toSend))
				delete(messageStore, messageType)
				messageKeys = messageKeys[1:]
			}
		case m := <-messages:
			fmt.Printf("Received message of type %d with message %s", m.Type, m.Message)
			exists := false
			for _, k := range messageKeys {
				if k == m.Type {
					exists = true
				}
			}
			if !exists {
				messageKeys = append(messageKeys, m.Type)
			}
			fmt.Printf("Before append: %+v\n", messageStore[m.Type])
			messageStore[m.Type] = append(messageStore[m.Type], m)
			fmt.Printf("After append: %+v\n", messageStore[m.Type])
		}
	}
}

func formatLottoMessage(t int, msgs []LottoMessage) string {
	fmt.Println(msgs)
	switch t {
	case End:
		return msgs[0].Message
	case Start:
		return msgs[0].Message
	case Error:
		return msgs[0].Message
	case Join:
		if len(msgs) == 1 {
			return fmt.Sprintf("%s has joined the lotto!", msgs[0].Message)
		} else {
			var buffer bytes.Buffer
			for _, msg := range msgs[:len(msgs)-2] {
				buffer.WriteString(msg.Message)
				buffer.WriteString(", ")
			}
			buffer.WriteString(msgs[len(msgs)-2].Message)
			buffer.WriteString(" and ")
			buffer.WriteString(msgs[len(msgs)-1].Message)
			buffer.WriteString(" have joined the lotto!")
			return buffer.String()
		}
	case Leave:
		if len(msgs) == 1 {
			return fmt.Sprintf("%s has been removed the lotto!", msgs[0].Message)
		} else {
			var buffer bytes.Buffer
			for _, msg := range msgs[:len(msgs)-2] {
				buffer.WriteString(msg.Message)
				buffer.WriteString(", ")
			}
			buffer.WriteString(msgs[len(msgs)-2].Message)
			buffer.WriteString(" and ")
			buffer.WriteString(msgs[len(msgs)-1].Message)
			buffer.WriteString(" have been removed from the lotto")
			return buffer.String()
		}
	}
	return ""
}

type Nick struct {
	Nick, Host, User string
}

type Lotto struct {
	Entries []Nick
	Host *Nick
	Prize string
	MessageChan chan LottoMessage
	CloseChan chan int
}

type LottoMessage struct {
	Type int
	Message string
}

func LottoNickHandler(lottos map[string]*Lotto) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		for k := range lottos {
			lottos[k].ChangeNick(line.Nick, line.Args[0])
		}
	}
}

func LottoPartHandler(lottos map[string]*Lotto) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		lotto, ok := lottos[line.Target()]
		if ok {
			if *lotto.Host == sender(line) {
				lottos[line.Target()].MessageChan <- LottoMessage{End,  "The lotto has been canceled because the host left"}
				lottos[line.Target()].Close()
				delete(lottos, line.Target())
				return
			}
			removedSomeone := lotto.Remove(sender(line))
			if removedSomeone {
				lotto.MessageChan <- LottoMessage{Leave, line.Nick}
			}
			// conn.Privmsgf(line.Target(), "%s have been removed from the lotto", line.Nick)
		}
	}
}

func LottoQuitHandler(lottos map[string]*Lotto) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		for k := range lottos {
			if *lottos[k].Host == sender(line) {
				lottos[k].MessageChan <- LottoMessage{End,  "The lotto has been canceled because the host left"}
				lottos[k].Close()
				delete(lottos, k)
				// conn.Privmsg(k, "The lotto has been canceled because the host left")
				return
			}
			removedSomeone := lottos[k].Remove(sender(line))
			if removedSomeone {
				lottos[k].MessageChan <- LottoMessage{Leave, line.Nick}
			}
			// conn.Privmsgf(k, "%s have been removed from the lotto", line.Nick)
		}
	}
}


func LottoPrivmsgHandler(lottos map[string]*Lotto) irc.HandlerFunc {
	return func(conn *irc.Conn, line *irc.Line) {
		sanitizeLottos(lottos, line.Target())
		isStartLotto := regexp.MustCompile(`^!startlotto\s+(.+)`)
		matches := isStartLotto.FindStringSubmatch(line.Text())
		if len(matches) > 1 {
			fmt.Println("Starting lotto")
			handleStartLotto(lottos[line.Target()], conn, line, matches[1])
		}
		isJoinLotto := regexp.MustCompile(`^!joinlotto(\s+.*|$)`)
		if len(isJoinLotto.FindStringSubmatch(line.Text())) > 0 {
			fmt.Println("Joining lotto")
			handleJoinLotto(lottos[line.Target()], conn, line)
		}
		isEndLotto := regexp.MustCompile(`^!endlotto(\s+.*|$)`)
		if len(isEndLotto.FindStringSubmatch(line.Text())) > 0 {
			fmt.Println("Ending lotto")
			ended := handleEndLotto(lottos[line.Target()], conn, line)
			if ended {
				delete(lottos, line.Target())
			}
		}
	}
}

func sanitizeLottos(lottos map[string]*Lotto, channel string) {
	for k := range lottos {
		if lottos[k].Host == nil {
			delete(lottos, k)
		}
	}
	_, ok := lottos[channel]
	if !ok {
		lottos[channel] = new(Lotto)
	}
}

func sender(line *irc.Line) Nick {
	return Nick{ line.Nick, line.Host, line.Ident }
}

func handleStartLotto(lotto *Lotto, conn *irc.Conn, line *irc.Line, prize string) {
	if lotto.Host != nil {
		lotto.MessageChan <- LottoMessage{Error,  "You cannot start a lotto while one is still running"}
		// conn.Privmsg(line.Target(), "You cannot start a lotto while one is still running")
		return
	}
	lotto.Start(sender(line), prize)
	go ProcessLottoMessages(conn, line.Target(), lotto.MessageChan, lotto.CloseChan)
	lotto.MessageChan <- LottoMessage{Start, fmt.Sprintf("%s has started a lotto for a %s!", lotto.Host.Nick, prize)}
	// conn.Privmsgf(line.Target(), "%s has started a lotto for a %s!", lotto.Host.Nick, prize)
}

func handleJoinLotto(lotto *Lotto, conn *irc.Conn, line *irc.Line) {
	if lotto.Host == nil {
		conn.Privmsg(line.Target(), "You cannot join that which does not exist")
		return
	}
	joined := lotto.Join(sender(line))
	if joined {
		lotto.MessageChan <- LottoMessage{Join, line.Nick}
		// conn.Privmsgf(line.Target(), "%s has joined the lotto!", line.Nick)
	} else {
		lotto.MessageChan <- LottoMessage{Error, fmt.Sprintf("%s has already joined the lotto", line.Nick)}
		// conn.Privmsgf(line.Target(), "%s has already joined the lotto", line.Nick)
	}
}

func handleEndLotto(lotto *Lotto, conn *irc.Conn, line *irc.Line) bool {
	if sameNick(sender(line), *lotto.Host) {
		winner := lotto.Winner()
		if winner == nil {
			lotto.MessageChan <- LottoMessage{End,  "Your empty lotto has ended without a winner. How sad!"}
			// conn.Privmsg(line.Target(), "Your empty lotto has ended without a winner. How sad!")
		} else {
			lotto.MessageChan <- LottoMessage{End, fmt.Sprintf("%s has won a %s!", winner.Nick, lotto.Prize)}
			// conn.Privmsgf(line.Target(), "%s has won a %s!", winner.Nick, lotto.Prize)
		}
		return true
	}
	lotto.MessageChan <- LottoMessage{Error, "Only the host of a lotto can end it!"}
	// conn.Privmsg(line.Target(), "Only the host of a lotto can end it!")
	return false
}

func sameNick(nick1 Nick, nick2 Nick) bool {
	return nick1.Host == nick2.Host
}

func (lotto *Lotto) Join(nick Nick) bool {
	for _, existingNick := range lotto.Entries {
		if sameNick(existingNick, nick) {
			return false
		}
	}
	lotto.Entries = append(lotto.Entries, nick)
	return true
}

func (lotto *Lotto) ChangeNick(from string, to string) {
	for i, nick := range lotto.Entries {
		if nick.Nick == from {
			lotto.Entries[i].Nick = to
		}
	}
	if lotto.Host.Nick == from {
		lotto.Host.Nick = to
	}
}

func (lotto *Lotto) Remove(nick Nick) bool {
	for i, enteredNick := range lotto.Entries {
		if sameNick(nick, enteredNick) {
			lotto.Entries = append(lotto.Entries[:i], lotto.Entries[i+1:]...)
			return true
		}
	}
	return false
}

func (lotto *Lotto) Start(host Nick, prize string) {
	lotto.Host = &host
	lotto.Prize = prize
	lotto.MessageChan = make(chan LottoMessage, 20)
	lotto.CloseChan = make(chan int, 1)
}

func (lotto *Lotto) Winner() *Nick {
	if len(lotto.Entries) == 0 {
		return nil
	}
	rand.Seed(time.Now().Unix())
	winnerIndex := rand.Intn(len(lotto.Entries))
	return &lotto.Entries[winnerIndex]
}

func (lotto *Lotto) Close() {
	lotto.CloseChan <- 0
}
