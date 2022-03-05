package TwitchCore

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

type TwitchCore struct {
	clientId    string
	accessToken string
}

func (twitchCore *TwitchCore) Initialize(clientId string, clientSecret string) error {
	twitchCore.clientId = clientId

	url := "https://id.twitch.tv/oauth2/token?client_id=" + twitchCore.clientId + "&client_secret=" + clientSecret + "&grant_type=client_credentials"
	reqBody := bytes.NewBufferString("Post")

	resp, err := http.Post(url, "", reqBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	twitchCore.accessToken = string(data)[17:47]
	return nil
}

func (twitchCore *TwitchCore) IsStreamerLive(streamerId string) bool {
	url := "https://api.twitch.tv/helix/search/channels?query=" + streamerId
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("client-id", twitchCore.clientId)
	req.Header.Add("Authorization", "Bearer "+twitchCore.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes)

	//fmt.Println(str)
	str = str[strings.Index(str, "\""+streamerId+"\""):]

	if strings.Contains(str, "is_live") {
		pos := strings.Index(str, "is_live") + 9
		str = str[pos : pos+5]

		if strings.Contains(str, "true") {
			return true
		}
	}
	return false
}
