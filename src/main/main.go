package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Keytg ключ бота Facebook
var Keyfb string

// Keytg ключ сервиса на yandex
var Keyyandex string

//  WebhookURL url сервера бота
const WebhookURL = "https://app2-test48.herokuapp.com/webhook"

var AccessToken string
var VerifyToken string
var Port string

//const FacebookEndPoint = "https://fb.me/Pprisnbot"
const FacebookEndPoint = "https://m.me/Pprisnbot"

// WebTranslateURL url сервиса переводчика на русский с английского
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"

//Структура данных ответа от https://translate.yandex.net/api/v1.5/tr.json/translate
// Пример на запрос Hellow Jack, ответ: {"code":200,"lang":"en-ru","text":["\"Хеллоу Джек\""]}
type TranslateJoke struct {
	CODE uint32   `json: "code"`
	Lang string   `json: "lang"`
	Text []string `json: "text"`
}

//Формат события Webhook
//{
//	"object":"page",
//	"entry":[
//	  {
//		"id":"<PAGE_ID>",
//		"time":1458692752478,
//		"messaging":[
//		  {
//			"sender":{
//			  "id":"<PSID>"
//			},
//			"recipient":{
//			  "id":"<PAGE_ID>"
//			},
//
//			...
//		  }
//		]
//	  }
//	]
// }
type ReceivedMessage struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID        int64       `json:"id"`
	Time      int64       `json:"time"`
	Messaging []Messaging `json:"messaging"`
}

type Messaging struct {
	Sender    Sender    `json:"sender"`
	Recipient Recipient `json:"recipient"`
	Timestamp int64     `json:"timestamp"`
	Message   Message   `json:"message"`
}

type Sender struct {
	ID int64 `json:"id"`
}

type Recipient struct {
	ID int64 `json:"id"`
}

type Message struct {
	MID  string `json:"mid"`
	Seq  int64  `json:"seq"`
	Text string `json:"text"`
}

type Payload struct {
	TemplateType string  `json:"template_type"`
	Text         string  `json:"text"`
	Buttons      Buttons `json:"buttons"`
}

type Buttons struct {
	Type  string `json:"type"`
	Url   string `json:"url"`
	Title string `json:"title"`
}

type Attachment struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type ButtonMessageBody struct {
	Attachment Attachment `json:"attachment"`
}

type ButtonMessage struct {
	Recipient         Recipient         `json:"recipient"`
	ButtonMessageBody ButtonMessageBody `json:"message"`
}

type SendMessage struct {
	Recipient Recipient `json:"recipient"`
	Message   struct {
		Text string `json:"text"`
	} `json:"message"`
}

//Проинициализируем ключи Keyfb Keyyandex
func init() {
	Port = os.Getenv("PORT")
	Keyfb = os.Getenv("KEYFB")
	Keyyandex = os.Getenv("KEYYANDEX")
	AccessToken = os.Getenv("ACCESS_TOKEN")
	VerifyToken = os.Getenv("VERIFY_TOKEN")
}

func webhookEndpoint(w http.ResponseWriter, r *http.Request) {
	//	fmt.Fprintln(w, "Hello :)")
	if r.Method == "GET" {
		verifyTokenAction(w, r)
	}
	if r.Method == "POST" {
		webhookPostAction(w, r)
	}
}

//curl -X GET "localhost:1337/webhook?hub.verify_token=<YOUR_VERIFY_TOKEN>&hub.challenge=CHALLENGE_ACCEPTED&hub.mode=subscribe"
func verifyTokenAction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("hub.verify_token") == VerifyToken {
		log.Print("verify token success.")
		fmt.Fprintf(w, r.URL.Query().Get("hub.challenge"))
	} else {
		log.Print("Error: verify token failed.")
		fmt.Fprintf(w, "Error, wrong validation token")
	}
}

//curl -H "Content-Type: application/json" -X POST "localhost:1337/webhook" -d '{"object": "page", "entry": [{"messaging": [{"message": "TEST_MESSAGE"}]}]}'
func webhookPostAction(w http.ResponseWriter, r *http.Request) {
	var receivedMessage ReceivedMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
	}
	if err = json.Unmarshal(body, &receivedMessage); err != nil {
		log.Print(err)
	}
	messagingEvents := receivedMessage.Entry[0].Messaging
	for _, event := range messagingEvents {
		senderID := event.Sender.ID

		log.Print("senderID: " + strconv.FormatInt(senderID, 10))
		log.Print("%+v", event)
		if &event.Message != nil && event.Message.Text != "" {
			// TODO: Fix sendButtonMessage function
			//if messageForButton(event.Message.Text) {
			//	message := getReplyMessage(event.Message.Text)
			//	sendButtonMessage(senderID, message)
			//} else {
			//	message := getReplyMessage(event.Message.Text)
			//	sendTextMessage(senderID, message)
			//}
			message := getReplyMessage(event.Message.Text)
			sendTextMessage(senderID, message)
		}
	}
	fmt.Fprintf(w, "Success")
}

// TODO: Reply message is just sample and made by easy logic, need to enhance the logic.
func getReplyMessage(receivedMessage string) string {
	var message string
	receivedMessage = strings.ToUpper(receivedMessage)
	log.Print(" Received message: " + receivedMessage)

	if strings.Contains(receivedMessage, "TEST") {
		message = "Вы запросили TEST"
	} else if strings.Contains(receivedMessage, "TEST1") {
		message = "Вы запросили TEST1"
	} else {
		message = " Ваше сообщение принято"
	}
	return message
}

func sendTextMessage(senderID int64, text string) {
	//	let request_body = {
	//		"recipient": {
	//		  "id": sender_psid
	//		},
	//		"message": response
	//	  }
	//	  // Send the HTTP request to the Messenger Platform
	//	  request({
	//		"uri": "https://graph.facebook.com/v2.6/me/messages",
	//		"qs": { "access_token": PAGE_ACCESS_TOKEN },
	//		"method": "POST",
	//		"json": request_body

	recipient := new(Recipient)
	recipient.ID = senderID
	send_message := new(SendMessage)
	send_message.Recipient = *recipient
	send_message.Message.Text = text
	send_message_body, err := json.Marshal(send_message)
	if err != nil {
		log.Print(err)
	}

	//var dialTimeout = time.Duration(30 * time.Second)
	//httpClient := &http.Client{
	//	Timeout: dialTimeout,
	//	Transport: &http.Transport{
	//		TLSClientConfig: &tls.Config{
	//			InsecureSkipVerify: true,
	//		},
	//	},
	//}

	//	//httpClient := new(http.Client)
	//res, err := httpClient.Post(FacebookEndPoint, `application/json; charset=utf-8;access_token=`+AccessToken, bytes.NewBuffer(send_message_body))

	req, err := http.NewRequest("POST", FacebookEndPoint, bytes.NewBuffer(send_message_body))
	if err != nil {
		log.Print(err)
	}
	fmt.Println("%+v", req)
	fmt.Println("%+v", err)

	values := url.Values{}
	values.Add("access_token", AccessToken)
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Duration(30 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}

	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/webhook", webhookEndpoint)
	if err := http.ListenAndServe(":"+Port, r); err != nil {
		log.Fatal(err)
	}
}
