package main

import (
	"crypto/tls"
	"flag"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

func main() {
	server := flag.String("s", "irc.example.net:6697", "IRC Server")
	nick := flag.String("n", "boat", "Nickname")
	user := flag.String("u", "boat", "Username")
	channels := flag.String("c", "#example1,#example2", "Comma separated list of channels to join")
	notUseTls := flag.Bool("xxx", false, "Do not use TLS")
	flag.Parse()

	runIrc(*server, *nick, *user, *notUseTls, strings.Split(*channels, ","))
}

func runIrc(server, nick, owner string, notUseTls bool, channels []string) {
	io := irc.IRC(nick, owner)
	io.UseTLS = !notUseTls
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
