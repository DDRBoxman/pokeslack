package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

var rares = []int{
	3,   //venasaur
	6,   //charizard
	9,   //blastoise
	64,  //kadabra
	65,  //alakazam
	113, //chansey
	138, //omanyte
	143, //snorlax
	149, //dragonite
}

type pokemon struct {
	Type    string  `json:"type"`
	Message message `json:"message"`
}

type message struct {
	PokemonID int     `json:"pokemon_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type slackmessage struct {
	Text        string       `json:"text"`
	IconURL     string       `json:"icon_url"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	Fallback string `json:"fallback"`
	Pretext  string `json:"pretext"`
	ImageURL string `json:"image_url"`
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

	log.Println(p)

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

func sendMessage(lat, lng float64, pokeID int) {
	mapURL := generateMap(lat, lng, pokeID)

	message := slackmessage{
		Text:        "Poke found!",
		UnfurlLinks: true,
		IconURL:     getPokeIconURL(pokeID),
		Attachments: []attachment{
			attachment{
				Fallback: "Poke found!",
				ImageURL: mapURL,
			},
		},
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

func generateMap(lat, lng float64, pokeID int) string {
	mapURL, err := url.Parse("https://maps.googleapis.com/maps/api/staticmap")
	if err != nil {
		return ""
	}

	q := mapURL.Query()
	q.Set("zoom", "15")
	q.Set("key", viper.GetString("GOOGLE_MAPS_KEY"))
	q.Set("center", fmt.Sprintf("%f,%f", lat, lng))
	q.Set("size", "400x400")
	q.Set("markers", fmt.Sprintf("icon:%s|%f,%f", getPokeIconURL(pokeID), lat, lng))
	mapURL.RawQuery = q.Encode()

	// warm up image cache since slack gives up really fast
	http.Get(mapURL.String())

	return mapURL.String()
}

func getPokeIconURL(pokeID int) string {
	return fmt.Sprintf("%sstatic/icons/%d.png", viper.GetString("POKEMAP_SERVER_URL"), pokeID)
}
