package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//https://mholt.github.io/json-to-go/
type LookupShow struct {
	ID        int      `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	Language  string   `json:"language"`
	Genres    []string `json:"genres"`
	Status    string   `json:"status"`
	Runtime   int      `json:"runtime"`
	Premiered string   `json:"premiered"`
	Schedule  struct {
		Time string   `json:"time"`
		Days []string `json:"days"`
	} `json:"schedule"`
	Rating struct {
		Average float64 `json:"average"`
	} `json:"rating"`
	Weight  int `json:"weight"`
	Network struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Country struct {
			Name     string `json:"name"`
			Code     string `json:"code"`
			Timezone string `json:"timezone"`
		} `json:"country"`
	} `json:"network"`
	WebChannel interface{} `json:"webChannel"`
	Externals  struct {
		Tvrage  int    `json:"tvrage"`
		Thetvdb int    `json:"thetvdb"`
		Imdb    string `json:"imdb"`
	} `json:"externals"`
	Image struct {
		Medium   string `json:"medium"`
		Original string `json:"original"`
	} `json:"image"`
	Summary string `json:"summary"`
	Updated int    `json:"updated"`
	Links   struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Previousepisode struct {
			Href string `json:"href"`
		} `json:"previousepisode"`
		Nextepisode struct {
			Href string `json:"href"`
		} `json:"nextepisode"`
	} `json:"_links"`
}

type EpisodeInfo struct {
	ID       int         `json:"id"`
	URL      string      `json:"url"`
	Name     string      `json:"name"`
	Season   int         `json:"season"`
	Number   int         `json:"number"`
	Airdate  string      `json:"airdate"`
	Airtime  string      `json:"airtime"`
	Airstamp string      `json:"airstamp"`
	Runtime  int         `json:"runtime"`
	Image    interface{} `json:"image"`
	Summary  string      `json:"summary"`
	Links    struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

func nextEpisode(series string) (string, error) {
	// try prev if this fails
	if r, e := episodeInfo(series, "is", func(s LookupShow) string { return s.Links.Nextepisode.Href }); e == nil {
		return r, nil
	}
	return lastEpisode(series)
}

func lastEpisode(series string) (string, error) {
	return episodeInfo(series, "was", func(s LookupShow) string { return s.Links.Previousepisode.Href })
}

func episodeInfo(series, verb string, selector func(LookupShow) string) (string, error) {
	show, e1 := lookupShow(series)
	if e1 != nil {
		return "", fmt.Errorf("No result for user input \"%s\" (%v)", series, e1)
	}

	episode, e2 := lookupEpisode(selector(show))
	if e2 != nil {
		return "", fmt.Errorf("No result for \"%s\" (%v)", show.Name, e2)
	}

	return fmt.Sprintf("\"%s\" %s on %s", show.Name, verb, parseTime(episode.Airstamp)), nil
}

func parseTime(airstamp string) string {
	l, e1 := time.LoadLocation("America/New_York")
	if e1 != nil {
		return airstamp
	}
	t, e2 := time.Parse(time.RFC3339, airstamp)
	if e2 != nil {
		return airstamp
	}
	return t.In(l).Format(time.UnixDate)
}

func lookupShow(series string) (LookupShow, error) {
	url := fmt.Sprintf("http://api.tvmaze.com/singlesearch/shows?q=%s", series)
	var f LookupShow
	if err := getJson(url, &f); err != nil {
		return f, err
	}
	return f, nil
}

func lookupEpisode(url string) (EpisodeInfo, error) {
	var e EpisodeInfo
	if err := getJson(url, &e); err != nil {
		return e, err
	}
	return e, nil
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
