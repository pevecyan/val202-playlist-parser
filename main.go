package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

const streamUrl = "http://mp3.rtvslo.si:8000/val202"

func main() {
	streamTitle, err := GetStreamTitle(streamUrl)
	if err != nil {
		panic(errors.Wrap(err, "get stream title"))
	}

	jsonData, err := json.Marshal(map[string]string{"title": streamTitle})
	if err != nil {
		panic(errors.Wrap(err, "marshal json"))
	}

	err = os.WriteFile("last_song.json", jsonData, 0644)
	if err != nil {
		panic(errors.Wrap(err, "write file"))
	}
}

func GetStreamTitle(streamUrl string) (string, error) {
	m, err := getStreamMetas(streamUrl)

	if err != nil {
		return "", err
	}
	// Should be at least "StreamTitle=' '"
	if len(m) < 15 {
		return "", nil
	}
	// Split meta by ';', trim it and search for StreamTitle
	for _, m := range bytes.Split(m, []byte(";")) {
		m = bytes.Trim(m, " \t")
		if bytes.Compare(m[0:13], []byte("StreamTitle='")) != 0 {
			continue
		}
		return string(m[13 : len(m)-1]), nil
	}
	return "", nil
}

func getStreamMetas(streamUrl string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", streamUrl, nil)
	req.Header.Set("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// We sent "Icy-MetaData", we should have a "icy-metaint" in return
	ih := resp.Header.Get("icy-metaint")
	if ih == "" {
		return nil, fmt.Errorf("no metadata")
	}
	// "icy-metaint" is how often (in bytes) should we receive the meta
	ib, err := strconv.Atoi(ih)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(resp.Body)

	// skip the first mp3 frame
	c, err := reader.Discard(ib)
	if err != nil {
		return nil, err
	}
	// If we didn't received ib bytes, the stream is ended
	if c != ib {
		return nil, fmt.Errorf("stream ended prematurally")
	}

	// get the size byte, that is the metadata length in bytes / 16
	sb, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	ms := int(sb * 16)

	// read the ms first bytes it will contain metadata
	m, err := reader.Peek(ms)
	if err != nil {
		return nil, err
	}
	return m, nil
}
