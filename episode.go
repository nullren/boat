package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Episode struct {
	Title        string
	Release_date string
	Season       int
	Number       int
	Show         *interface{}
}

type EpGuideResponse struct {
	Episode *Episode
}

func nextEpisode(series string) (string, error) {
	return epguide("next", series)
}

func lastEpisode(series string) (string, error) {
	return epguide("last", series)
}

func epguide(cmd, series string) (string, error) {
	url := fmt.Sprintf("%s/%s/", epguideUrl(series), cmd)

	var f EpGuideResponse
	err := getJson(url, &f)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("s%02de%02d: %s", f.Episode.Season, f.Episode.Number, f.Episode.Release_date), nil
}

func stripWs(something string) string {
	return strings.Join(strings.Fields(something), "")
}

func epguideUrl(series string) string {
	return "https://epguides.frecar.no/show/" + stripWs(series)
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func failif(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
