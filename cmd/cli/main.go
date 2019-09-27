package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/golang/protobuf/proto"
	gonats "github.com/nats-io/go-nats"
	senderNats "github.com/void616/gm-mint-sender/pkg/sender/nats/sender"
	watcherNats "github.com/void616/gm-mint-sender/pkg/watcher/nats/wallet"
	sumuslib "github.com/void616/gm-sumuslib"
	"github.com/void616/gm-sumuslib/amount"
)

var (
	nats           *gonats.Conn
	natsSubjPrefix *string
	tags           = make(map[string]bool)
	tagsLock       sync.Mutex
)

func main() {

	// flags
	natsURL := flag.String("nats", "localhost:4222", "Nats server endpoint")
	natsSubjPrefix = flag.String("nats-prefix", "", "Prefix for Nats messages subject")
	flag.Parse()

	// nats prefix: add dot
	if *natsSubjPrefix != "" && !strings.HasSuffix(*natsSubjPrefix, ".") {
		*natsSubjPrefix = *natsSubjPrefix + "."
	}

	// nats connection
	{
		nc, err := gonats.Connect(
			*natsURL,
			gonats.MaxReconnects(-1),
		)
		if err != nil {
			failln("Failed to connect to Nats server: %v", err)
			os.Exit(1)
		}
		nats = nc

		nats.SetDisconnectHandler(func(_ *gonats.Conn) {
			event("Nats disconnected")
		})
		nats.SetReconnectHandler(func(_ *gonats.Conn) {
			event("Nats connected")
		})
	}

	wg := sync.WaitGroup{}
	stopped := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		natsSubscribeRefillings()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		natsSubscribeSendings()
	}()

	// read input
	echo("Type ")
	success("help ")
	echo("to get help or ")
	success("exit ")
	echo("to exit.\n")
	input := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := input.ReadString('\n')
		if err != nil {
			panic(fmt.Sprintf("Failed to parse command: %v", err))
		}
		line = strings.TrimSpace(line)

		// help or exit
		switch {
		case line == "":
			continue
		case line == "help" || line == "?":
			(&cmdAddRemoveWallet{}).Help()
			(&cmdSend{}).Help()
			continue
		case line == "exit":
			echoln("Bye!")
			nats.Drain()
			close(stopped)
			wg.Wait()
			return
		}

		// command
		var cmd command
		switch {
		case (&cmdAddRemoveWallet{}).Is(line):
			cmd = &cmdAddRemoveWallet{}
		case (&cmdSend{}).Is(line):
			cmd = &cmdSend{}
		default:
			failln("Unknown command: %v", line)
			continue
		}

		// parse and perform
		if err := cmd.Parse(line); err != nil {
			failln("Parsing error: %v", err)
			continue
		}
		msg, err := cmd.Perform()
		if err != nil {
			failln("Performing error: %v", err)
			continue
		}
		successln(msg)
	}
}

// ---
// Nats subscriptions
// ---

func natsSubscribeRefillings() {
	subj := *natsSubjPrefix + watcherNats.SubjectRefill
	_, err := nats.Subscribe(subj, func(m *gonats.Msg) {
		// get
		reqModel := watcherNats.RefillEvent{}
		if err := proto.Unmarshal(m.Data, &reqModel); err != nil {
			failln("Failed to unmarshal: %v", err)
			return
		}
		// check service
		if !hasTag(reqModel.GetService()) {
			event("Deposit tag %v ignored (wallet %v)", reqModel.GetService(), reqModel.GetPublicKey())
			return
		}
		// reply
		repModel := watcherNats.RefillEventReply{Success: true}
		rep, err := proto.Marshal(&repModel)
		if err != nil {
			failln("Failed to marshal: %v", err)
			return
		}
		if err := nats.Publish(m.Reply, rep); err != nil {
			failln("Failed to reply: %v", err)
			return
		}
		event("%v %v deposited to %v, tag %v, tx %v", reqModel.GetAmount(), reqModel.GetToken(), reqModel.GetPublicKey(), reqModel.GetService(), reqModel.GetTransaction())
	})
	if err != nil {
		failln("Failed to subscribe: %v", err)
		return
	}
}

func natsSubscribeSendings() {
	subj := *natsSubjPrefix + senderNats.SubjectSent
	_, err := nats.Subscribe(subj, func(m *gonats.Msg) {
		// get
		reqModel := senderNats.SentEvent{}
		if err := proto.Unmarshal(m.Data, &reqModel); err != nil {
			failln("Failed to unmarshal: %v", err)
			return
		}
		// check service
		if !hasTag(reqModel.GetService()) {
			event("Witdrawal tag %v ignored", reqModel.GetService())
			return
		}
		// reply
		repModel := senderNats.SentEventReply{Success: true}
		rep, err := proto.Marshal(&repModel)
		if err != nil {
			failln("Failed to marshal: %v", err)
			return
		}
		if err := nats.Publish(m.Reply, rep); err != nil {
			failln("Failed to reply: %v", err)
			return
		}
		if reqModel.GetSuccess() {
			event("Witdrawal #%v (%v %v to %v) completed, tag %v, tx %v", reqModel.GetId(), reqModel.GetAmount(), reqModel.GetToken(), sumuslib.MaskString6P4(reqModel.GetPublicKey()), reqModel.GetService(), reqModel.GetTransaction())
		} else {
			event("Witdrawal #%v failed, tag %v. Error: %v", reqModel.GetId(), reqModel.GetService(), reqModel.GetError())
		}
	})
	if err != nil {
		failln("Failed to subscribe: %v", err)
		return
	}
}

// ---
// Commands
// ---

type command interface {
	Is(string) bool
	Parse(string) error
	Perform() (string, error)
	Help()
}

type cmdAddRemoveWallet struct {
	add    bool
	pubkey string
	tag    string
}

func (c *cmdAddRemoveWallet) Is(s string) bool {
	return strings.HasPrefix(s, "watch ") || strings.HasPrefix(s, "unwatch ")
}

func (c *cmdAddRemoveWallet) Parse(s string) error {
	var action string
	if _, err := fmt.Sscanf(s, "%s %s %s", &action, &c.pubkey, &c.tag); err != nil {
		return err
	}
	if _, err := sumuslib.ParsePublicKey(c.pubkey); err != nil {
		return err
	}
	c.add = action == "watch"
	return nil
}

func (c *cmdAddRemoveWallet) Perform() (string, error) {
	req, _ := proto.Marshal(&watcherNats.AddRemoveRequest{
		Service:   c.tag,
		Add:       c.add,
		PublicKey: []string{c.pubkey},
	})
	msg, err := nats.Request(*natsSubjPrefix+watcherNats.SubjectWatch, req, time.Second*5)
	if err != nil || msg == nil {
		return "", fmt.Errorf("send request: %v", err)
	}
	rep := watcherNats.AddRemoveReply{}
	if err := proto.Unmarshal(msg.Data, &rep); err != nil {
		return "", fmt.Errorf("unmarshal: %v", err)
	}
	if rep.GetSuccess() {
		watchTag(c.tag, c.add)
		if c.add {
			return fmt.Sprintf("Done. Added (tag %v)", c.tag), nil
		}
		return fmt.Sprintf("Done. Removed (tag %v)", c.tag), nil
	}
	return "", fmt.Errorf("service error: %v", rep.GetError())
}

func (c *cmdAddRemoveWallet) Help() {
	success("watch ")
	echo("<public_key> <tag> ")
	echoln("Add a wallet to the watcher-service")
	success("unwatch ")
	echo("<public_key> <tag> ")
	echoln("Remove a wallet from the watcher-service")
}

type cmdSend struct {
	amo    string
	token  string
	pubkey string
	tag    string
}

func (c *cmdSend) Is(s string) bool {
	return strings.HasPrefix(s, "send ")
}

func (c *cmdSend) Parse(s string) error {
	var null string
	if _, err := fmt.Sscanf(s, "%s %s %s %s %s", &null, &c.amo, &c.token, &c.pubkey, &c.tag); err != nil {
		return err
	}
	if _, err := sumuslib.ParseToken(c.token); err != nil {
		return err
	}
	if _, err := amount.FromString(c.amo); err != nil {
		return fmt.Errorf("failed to parse amount")
	}
	if _, err := sumuslib.ParsePublicKey(c.pubkey); err != nil {
		return err
	}
	return nil
}

func (c *cmdSend) Perform() (string, error) {
	id := fmt.Sprint(time.Now().UTC().UnixNano())
	req, _ := proto.Marshal(&senderNats.SendRequest{
		Service:   c.tag,
		Id:        id,
		PublicKey: c.pubkey,
		Amount:    c.amo,
		Token:     c.token,
	})
	msg, err := nats.Request(*natsSubjPrefix+senderNats.SubjectSend, req, time.Second*5)
	if err != nil || msg == nil {
		return "", fmt.Errorf("send request: %v", err)
	}
	rep := senderNats.SendReply{}
	if err := proto.Unmarshal(msg.Data, &rep); err != nil {
		return "", fmt.Errorf("unmarshal: %v", err)
	}
	if rep.GetSuccess() {
		watchTag(c.tag, true)
		return fmt.Sprintf("Done. Transfer %v (tag %v)", id, c.tag), nil
	}
	return "", fmt.Errorf("service error: %v", rep.GetError())
}

func (c *cmdSend) Help() {
	success("send ")
	echo("<amount> <gold|mnt> <public_key> <tag> ")
	echoln("Send a token transferring request to the sender-service")
}

// ---
// Helpers
// ---

func watchTag(t string, watch bool) {
	tagsLock.Lock()
	defer tagsLock.Unlock()
	tags[t] = watch
}

func hasTag(t string) bool {
	tagsLock.Lock()
	defer tagsLock.Unlock()
	if v, ok := tags[t]; v && ok {
		return true
	}
	return false
}

func echo(f string, args ...interface{}) {
	c := color.New(color.FgWhite)
	c.Printf(f, args...)
}

func echoln(f string, args ...interface{}) {
	echo(f+"\n", args...)
}

func event(f string, args ...interface{}) {
	c := color.New(color.FgYellow)
	c.Printf("\n"+f+"\n", args...)
}

func success(f string, args ...interface{}) {
	c := color.New(color.FgGreen)
	c.Printf(f, args...)
}

func successln(f string, args ...interface{}) {
	success(f+"\n", args...)
}

func fail(f string, args ...interface{}) {
	c := color.New(color.FgRed)
	c.Printf(f, args...)
}

func failln(f string, args ...interface{}) {
	fail(f+"\n", args...)
}
