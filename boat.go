package main

import (
	"crypto/tls"
	"flag"
	"github.com/thoj/go-ircevent"
	"log"
)

func main() {
	// load config options
	server := flag.String("server", "chat.freenode.net:6697", "Server to connect to")
	nick := flag.String("nick", "boat", "Nickname to use")
	user := flag.String("user", "boat", "Username to use")
	flag.Parse()
	server := "irc.example.net:6697"
	nick := "goboat"
	owner := "goboat"
	channels := []string{"#goboat"}

	runIrc(server, nick, owner, channels)
}

func runIrc(server, nick, owner string, channels []string) {
	io := irc.IRC(nick, owner)
	io.UseTLS = true
	io.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	io.Debug = true
	io.VerboseCallbackHandler = true

	io.AddCallback("001", func(e *irc.Event) {
		for _, c := range channels {
			io.Join(c)
		}
	})

	// hi responder
	io.AddCallback("PRIVMSG", func(event *irc.Event) {
		target := event.Arguments[0]
		if target == nick {
			target = event.Nick
		}

		if m := event.Message(); m == "hi" {
			io.Privmsgf(target, "hi, %v", event.Nick)
		}
	})

	// high five stealer
	io.AddCallback("PRIVMSG", func(event *irc.Event) {
		target := event.Arguments[0]
		if target == nick {
			target = event.Nick
		}

		if m := event.Message(); m == "o/" {
			io.Privmsgf(target, "%v: \\o", event.Nick)
		}
	})

	err := io.Connect(server)
	if err != nil {
		log.Fatal(err)
	}

	io.Loop()
}

func failiferr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
