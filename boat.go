package main

import (
	"crypto/tls"
	"flag"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nullren/go-ircevent"
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
	io.VerboseCallbackHandler = true

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
		go func() {
			for r := range reminders.Watch() {
				io.Privmsgf(r.Where, "%s: You asked me to remind you %s", r.Who, r.What)
				reminders.Save()
			}
		}()
		io.AddCallback("PRIVMSG", func(event *irc.Event) {
			target := event.Arguments[0]
			if target == nick {
				target = event.Nick
			}
			suffixTranslate := map[string]time.Duration{
				"s": time.Second,
				"m": time.Minute,
				"h": time.Hour,
				"d": time.Hour * 24,
				"w": time.Hour * 24 * 7,
			}
			cmd := "remind me in "
			if m := event.Message(); strings.HasPrefix(strings.ToLower(m), cmd) {
				meat := strings.Fields(m[len(cmd):])
				t := strings.ToLower(meat[0])
				r := strings.Join(meat[1:], " ")
				re := regexp.MustCompile("^[0-9]+[smhdw]$")
				if re.MatchString(t) {
					m, _ := strconv.Atoi(t[:len(t)-1])
					s := suffixTranslate[t[len(t)-1:]]
					when := time.Now().Add(s * time.Duration(m))
					reminders.Add(event.Nick, r, target, when)
					reminders.Save()
					io.Privmsgf(target, "%s: Okay, I'll remind you about that on %s", event.Nick, when.Format(time.RFC850))
				} else {
					io.Privmsgf(target, "%s: Woops. I only understand times in the form of \\d+[smhdw]", event.Nick)
				}
			}
		})
	}

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
