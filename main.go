package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

var rares = []string{
	"3",   //venasaur
	"6",   //charizard
	"9",   //blastoise
	"64",  //kadabra
	"65",  //alakazam
	"138", //omanyte
	"143", //snorlax
	"149", //dragonite
}

type pokemon struct {
	Type    string  `json:"type"`
	Message message `json:"type"`
}

type message struct {
	EncounterID   string `json:"encounter_id"`
	SpawnpointID  string `json:"spawnpoint_id"`
	PokemonID     string `json:"pokemon_id"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	DisappearTime string `json:"disappear_time"`
}

type slackmessage struct {
	Text        string `json:"text"`
	IconURL     string `json:"icon_url"`
	UnfurlLinks bool   `json:"unfurl_links"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	var p pokemon
	err := decoder.Decode(&p)
	if err != nil {
		log.Println(err)
		return
	}

	for _, v := range rares {
		if v == p.Message.PokemonID {
			sendMessage(p.Message.Latitude, p.Message.Longitude, p.Message.PokemonID)
		}
	}
}

func main() {
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.ReadInConfig()

	http.HandleFunc("/pokemon", handler)
	http.ListenAndServe(":9000", nil)
}

func sendMessage(lat, lng, pokeID string) {
	mapURL := generateMap(lat, lng, pokeID)

	message := slackmessage{
		Text:        fmt.Sprintf("Poke found!, <%s|Map>", mapURL),
		UnfurlLinks: true,
		IconURL:     getPokeIconURL(pokeID),
	}

	messageString, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = http.PostForm(viper.GetString("SLACK_WEBHOOK_URL"), url.Values{"payload": {string(messageString)}})
	if err != nil {
		log.Println(err)
	}
}

func generateMap(lat, lng, pokeID string) string {
	mapURL, err := url.Parse("https://maps.googleapis.com/maps/api/staticmap")
	if err != nil {
		return ""
	}

	q := mapURL.Query()
	q.Set("zoom", "15")
	q.Set("key", viper.GetString("GOOGLE_MAPS_KEY"))
	q.Set("center", lat+","+lng)
	q.Set("size", "400x400")
	q.Set("markers", fmt.Sprintf("icon:%s|%s,%s", getPokeIconURL(pokeID), lat, lng))
	mapURL.RawQuery = q.Encode()

	return mapURL.String()
}

func getPokeIconURL(pokeID string) string {
	return viper.GetString("POKEMAP_SERVER_URL") + "static/icons/" + pokeID + ".png"
}
