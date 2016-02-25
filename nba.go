package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	Date          time.Time
	VisitingTeam  string
	VisitingScore int
	HomeTeam      string
	HomeScore     int
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByTime []*Game

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

func atoi(i string) int {
	if i, err := strconv.Atoi(i); err == nil {
		return i
	} else {
		return -1
	}
}

func loadCsv(file string) ([][]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func sanitize(input string) string {
	r := regexp.MustCompile("[^A-Za-z ]+")
	return r.ReplaceAllLiteralString(
		strings.TrimSpace(strings.ToLower(input)), "")
}

func nbaLoadGames(file, team string) ([]*Game, error) {
	games := make([]*Game, 0)

	records, err := loadCsv(file)
	if err != nil {
		return games, err
	}

	const longForm = "Mon Jan 2 2006 3:04 pm"
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		return games, err
	}

	for _, record := range records {
		t, err := time.ParseInLocation(longForm, fmt.Sprintf("%s %s", record[0], record[1]), est)
		if err != nil {
			log.Printf("[WARNING] failed to parse record: %v (%v)\n", record, err)
			continue
		}
		game := &Game{
			Date:          t,
			VisitingTeam:  record[3],
			VisitingScore: atoi(record[4]),
			HomeTeam:      record[5],
			HomeScore:     atoi(record[6]),
		}
		v, _ := regexp.MatchString(sanitize(team), sanitize(game.VisitingTeam))
		h, _ := regexp.MatchString(sanitize(team), sanitize(game.HomeTeam))
		if v || h {
			games = append(games, game)
		}
	}
	return games, nil
}

func nbaLastGameBeforeTime(games []*Game, t time.Time) *Game {
	i := sort.Search(len(games), func(i int) bool { return games[i].Date.After(time.Now()) })
	return games[i-1]
}

func nbaNextGameAfterTime(games []*Game, t time.Time) *Game {
	i := sort.Search(len(games), func(i int) bool { return games[i].Date.After(time.Now()) })
	return games[i]
}

func nbaTest() {
	games, _ := nbaLoadGames("nba_2016.csv", "warriors")
	sort.Sort(ByTime(games))
	for _, game := range games {
		fmt.Println(game)
	}

	fmt.Println("Today's game")
	// get next game after today
	fmt.Println(nbaLastGameBeforeTime(games, time.Now()))
}
