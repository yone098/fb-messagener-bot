package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var channelID string
var channelSecret string
var channelMID string

const facebookPostURL = "https://graph.facebook.com/v2.6/me/messages?access_token="

var facebookVerifyToken string
var facebookToken string

type id struct {
	ID int `json:"id"`
}

type text struct {
	Text string `json:"text"`
}

type facebookMsg struct {
	Object string   `json:"object"`
	Entry  []*entry `json:"entry"`
}

type entry struct {
	ID        int64        `json:"id"`
	Time      int64        `json:"time"`
	Messaging []*messaging `json:"messaging"`
}

type messaging struct {
	Sender    *id       `json:"sender"`
	Recipient *id       `json:"recipient"`
	Timestamp int64     `json:"timestamp"`
	Message   *message  `json:"message"`
	Delivery  *delivery `json:"delivery"`
}

type message struct {
	Mid  string `json:"mid"`
	Seq  int64  `json:"seq"`
	Text string `json:"text"`
}

type delivery struct {
	Mids      []string `json:"mids"`
	Watermark int64    `json:"watermark"`
	Seq       int64    `json:"seq"`
}

type sendMessage struct {
	Recipient *id   `json:"recipient"`
	Message   *text `json:"message"`
}

func index(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Println("called index")
	fmt.Fprintf(w, "Hello %s!", "hoge")
}

func handleGetCallback(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Println("called callback GET")

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get parameter
	token := r.Form.Get("hub.verify_token")
	log.Println("hub.verify_token:", token)

	challenge := r.Form.Get("hub.challenge")
	log.Println("hub.challenge:", challenge)

	if token == facebookVerifyToken {
		fmt.Fprintf(w, challenge)
		return
	}
	fmt.Fprintf(w, "OK")
}

func handlePostCallback(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Println("recieve message from facebook messagenger")
	var msg facebookMsg
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("msg:", string(b))
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, event := range msg.Entry[0].Messaging {
		senderID := event.Sender.ID
		if event.Message != nil {
			log.Println("Recieved msg:", event.Message.Text)
			m := sendMessage{
				Recipient: &id{ID: senderID},
				Message:   &text{Text: "ハゲ"},
			}
			b, err := json.Marshal(m)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			req, err := http.NewRequest("POST", facebookPostURL+facebookToken, bytes.NewBuffer(b))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			req.Header.Add("Content-Type", "application/json; charset=UTF-8")
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var result map[string]interface{}
			err = json.Unmarshal(body, &result)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Println("Response: ", result)

		}
	}

	fmt.Fprintf(w, "OK")
}

// main function
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	flag.Set("bind", ":"+port)

	channelID = os.Getenv("LINE_BOT_CHANNEL_ID")
	log.Println("CHANNEL_ID:", channelID)
	channelSecret = os.Getenv("LINE_BOT_CHANNEL_SECRET")
	log.Println("CHANNEL_SECRET:", channelSecret)
	channelMID = os.Getenv("LINE_BOT_CHANNEL_MID")
	log.Println("CHANNEL_MID:", channelMID)

	facebookToken = os.Getenv("FACEBOOK_TOKEN")
	facebookVerifyToken = os.Getenv("FACEBOOK_VERIFY_TOKEN")

	goji.Get("/", index)
	goji.Get("/fb/callback", handleGetCallback)
	goji.Post("/fb/callback", handlePostCallback)
	goji.Serve()
}
