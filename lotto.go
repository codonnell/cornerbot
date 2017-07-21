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
			if m.Type == End || m.Type == Start || m.Type == Error {
				msgs := make([]LottoMessage, 1)
				msgs[0] = m
				conn.Privmsg(channel, formatLottoMessage(m.Type, msgs))
				break
			}
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
		return Orange(msgs[0].Message)
	case Start:
		return Orange(msgs[0].Message)
	case Error:
		return msgs[0].Message
	case Join:
		if len(msgs) == 1 {
			return Purple(fmt.Sprintf("%s grabs a cold one from the fridge for a chance to win the lotto.", msgs[0].Message))
		} else {
			var buffer bytes.Buffer
			for _, msg := range msgs[:len(msgs)-2] {
				buffer.WriteString(msg.Message)
				buffer.WriteString(", ")
			}
			buffer.WriteString(msgs[len(msgs)-2].Message)
			buffer.WriteString(" and ")
			buffer.WriteString(msgs[len(msgs)-1].Message)
			buffer.WriteString(" have grabbed a cold one from the fridge")
			return Purple(buffer.String())
		}
	case Leave:
		if len(msgs) == 1 {
			return fmt.Sprintf("%s drops their drink and it shatters. Party foul!", msgs[0].Message)
		} else {
			var buffer bytes.Buffer
			for _, msg := range msgs[:len(msgs)-2] {
				buffer.WriteString(msg.Message)
				buffer.WriteString(", ")
			}
			buffer.WriteString(msgs[len(msgs)-2].Message)
			buffer.WriteString(" and ")
			buffer.WriteString(msgs[len(msgs)-1].Message)
			buffer.WriteString(" simultaneously drop their drinks. What a show!")
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
		isStartLotto := regexp.MustCompile(`^!start\s+(.+)`)
		matches := isStartLotto.FindStringSubmatch(line.Text())
		if len(matches) > 1 {
			fmt.Println("Starting lotto")
			handleStartLotto(lottos[line.Target()], conn, line, matches[1])
		}
		isJoinLotto := regexp.MustCompile(`^!chill(\s+.*|$)`)
		if len(isJoinLotto.FindStringSubmatch(line.Text())) > 0 {
			fmt.Println("Joining lotto")
			handleJoinLotto(lottos[line.Target()], conn, line)
		}
		isEndLotto := regexp.MustCompile(`^!end(\s+.*|$)`)
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
		lotto.MessageChan <- LottoMessage{Error,  fmt.Sprintf("Silly %s. There's already a lotto running!", line.Nick)}
		// conn.Privmsg(line.Target(), "You cannot start a lotto while one is still running")
		return
	}
	senderMode := conn.StateTracker().GetChannel(line.Target()).Nicks[line.Nick]
	if senderMode.Owner || senderMode.Admin || senderMode.Op {
		lotto.Start(sender(line), prize)
		go ProcessLottoMessages(conn, line.Target(), lotto.MessageChan, lotto.CloseChan)
		lotto.MessageChan <- LottoMessage{Start, fmt.Sprintf("%s just brought a cooler full of refreshing beverages. Type !chill for a chance to win a/an %s.", lotto.Host.Nick, prize)}
	}
	// conn.Privmsgf(line.Target(), "%s has started a lotto for a %s!", lotto.Host.Nick, prize)
}

func handleJoinLotto(lotto *Lotto, conn *irc.Conn, line *irc.Line) {
	if lotto.Host == nil {
		conn.Privmsg(line.Target(), "There's no lotto running at the moment.")
		return
	}
	if line.Nick == lotto.Host.Nick {
		lotto.MessageChan <- LottoMessage{Error, "You can't join your own lotto, silly goose."}
		return
	}
	joined := lotto.Join(sender(line))
	if joined {
		lotto.MessageChan <- LottoMessage{Join, line.Nick}
		// conn.Privmsgf(line.Target(), "%s has joined the lotto!", line.Nick)
	} else {
		lotto.MessageChan <- LottoMessage{Error, fmt.Sprintf("Silly %s. One drink at a time.", line.Nick)}
		// conn.Privmsgf(line.Target(), "%s has already joined the lotto", line.Nick)
	}
}

func handleEndLotto(lotto *Lotto, conn *irc.Conn, line *irc.Line) bool {
	senderMode := conn.StateTracker().GetChannel(line.Target()).Nicks[line.Nick]
	if sameNick(sender(line), *lotto.Host) || senderMode.Owner || senderMode.Admin {
		winner := lotto.Winner()
		if winner == nil {
			lotto.MessageChan <- LottoMessage{End,  "No one wins, because no one had a drink! The lotto is over."}
			// conn.Privmsg(line.Target(), "Your empty lotto has ended without a winner. How sad!")
		} else {
			lotto.MessageChan <- LottoMessage{End, fmt.Sprintf("%s has the yummiest beverage and won a/an %s!", winner.Nick, lotto.Prize)}
			// conn.Privmsgf(line.Target(), "%s has won a %s!", winner.Nick, lotto.Prize)
		}
		return true
	}
	lotto.MessageChan <- LottoMessage{Error, "Only the host of a lotto can end it!"}
	// conn.Privmsg(line.Target(), "Only the host of a lotto can end it!")
	return false
}

func sameNick(nick1 Nick, nick2 Nick) bool {
	return nick1.User == nick2.User && nick1.Host == nick2.Host
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

func Orange(text string) string {
	var buffer bytes.Buffer
	buffer.WriteString("\x037")
	buffer.WriteString(text)
	buffer.WriteString("\x03")
	return buffer.String()
}

func Purple(text string) string {
	var buffer bytes.Buffer
	buffer.WriteString("\x036")
	buffer.WriteString(text)
	buffer.WriteString("\x03")
	return buffer.String()
}
