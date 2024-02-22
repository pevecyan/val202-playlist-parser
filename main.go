package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

const clientID = "change-me"

func main() {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	response, err := client.Get(fmt.Sprintf("https://api.rtvslo.si/preslikave/sos?station=val202&client_id=%s", clientID))
	if err != nil {
		panic(errors.Wrap(err, "get data"))
	}

	var apiResponse ApiResponse
	if err := json.NewDecoder(response.Body).Decode(&apiResponse); err != nil {
		panic(errors.Wrap(err, "decode response"))
	}

	lastItem := apiResponse.Response[len(apiResponse.Response)-1]

	jsonData, err := json.Marshal(lastItem)
	if err != nil {
		panic(errors.Wrap(err, "marshal json"))
	}

	err = os.WriteFile("last_song.json", jsonData, 0644)
	if err != nil {
		panic(errors.Wrap(err, "write file"))
	}
}

type ApiResponse struct {
	Response []PlaylistItem `json:"response"`
}

type PlaylistItem struct {
	StarTime   string `json:"start_time"`
	ArtistName string `json:"artist_name"`
	TitleName  string `json:"title_name"`
	Author     string `json:"author"`
}
