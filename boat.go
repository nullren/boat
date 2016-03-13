package main

import (
	"crypto/tls"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/nullren/go-ircevent"
	"github.com/nullren/goatparse"
)

func main() {
	server := flag.String("s", "irc.example.net:6697", "IRC Server")
	nick := flag.String("n", "boat", "Nickname")
	user := flag.String("u", "boat", "Username")
	password := flag.String("p", "boat", "Password")
	channels := flag.String("c", "#example1,#example2", "Comma separated list of channels to join")
	remindersFile := flag.String("r", "reminders.json", "Path to store reminders")
	useSasl := flag.Bool("sasl", false, "Use SASL")
	notUseTls := flag.Bool("insecure", false, "Do not use TLS")
	flag.Parse()

	runIrc(*server, *nick, *user, *password, *remindersFile, *useSasl, *notUseTls, strings.Split(*channels, ","))
}

func runIrc(server, nick, user, password, remindersFile string, useSasl, notUseTls bool, channels []string) {
	io := irc.IRC(nick, user)
	io.UseTLS = !notUseTls
	io.UseSasl = useSasl
	io.SaslPassword = password
	io.SaslUser = user
	io.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	io.Debug = true
	io.VerboseCallbackHandler = false

	io.AddCallback("001", func(e *irc.Event) {
		for _, c := range channels {
			io.Join(c)
		}
	})

	io.AddCallback("376", func(e *irc.Event) {
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

	// next episode
	io.AddCallback("PRIVMSG", func(event *irc.Event) {
		target := event.Arguments[0]
		if target == nick {
			target = event.Nick
		}
		cmd := ",next "
		if m := event.Message(); strings.HasPrefix(m, cmd) {
			res, err := nextEpisode(m[len(cmd):])
			if err == nil {
				io.Privmsgf(target, "%s", res)
			} else {
				io.Privmsgf(target, "Uh oh. %v", err)
			}
		}
	})

	// last episode
	io.AddCallback("PRIVMSG", func(event *irc.Event) {
		target := event.Arguments[0]
		if target == nick {
			target = event.Nick
		}
		cmd := ",last "
		if m := event.Message(); strings.HasPrefix(m, cmd) {
			res, err := lastEpisode(m[len(cmd):])
			if err == nil {
				io.Privmsgf(target, "%s", res)
			} else {
				io.Privmsgf(target, "Uh-oh. %v", err)
			}
		}
	})

	// remind me in...
	reminders, erem := InitializeReminders(remindersFile)
	if erem != nil {
		log.Printf("WARNING: Could not initialize reminders: %v\n", erem)
	} else {
		io.AddCallback("001", func(e *irc.Event) {
			time.Sleep(3 * time.Second) // who knows how long it takes to join channels
			go reminders.Watch(func(r *Reminder) {
				io.Privmsgf(r.Where, "%s: You asked me to remind you %s", r.Who, r.What)
				reminders.Save()
			})
		})
		io.AddCallback("PRIVMSG", func(event *irc.Event) {
			target := event.Arguments[0]
			if target == nick {
				target = event.Nick
			}
			cmd := "remind me in "
			if m := event.Message(); strings.HasPrefix(strings.ToLower(m), cmd) {
				pd, err := goatparse.ParseDuration(m[len(cmd):])
				if err == nil {
					when := pd.Time
					what := strings.TrimSpace(m[len(cmd)+pd.Offset:])
					reminders.Notify(reminders.Add(event.Nick, what, target, when))
					reminders.Save()
					io.Privmsgf(target, "%s: Okay, I'll remind you about that on %s", event.Nick, when.Format(time.RFC850))
				} else {
					io.Privmsgf(target, "%s: Woops, I didn't understand that", event.Nick)
				}
			}
		})
	}

	err := io.Connect(server)
	failif(err)

	io.Loop()
}

func failif(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
