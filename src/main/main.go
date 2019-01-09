package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
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

//структура данных ответа от api.icndb.com
// { "type": "success", "value": { "id": 563, "joke": "Chuck Norris causes the Windows Blue Screen of Death.", "categories": ["nerdy"] } }
// Описание вложенной структуры ответа ..{ "id": 563, "joke": "Chuck
type Joke struct {
	ID   uint32 `json: "id"`
	Joke string `json: "joke"`
}

// Описание начала структуры ответа  { "type": "success", "value": {..
type JokeResponse struct {
	Type  string `json:"type"`
	Value Joke   `json:"value"`
}

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
	ID        string      `json:"id"`
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
	ID string `json:"id"`
}

type Recipient struct {
	ID string `json:"id"`
}

type Message struct {
	MID    string `json:"mid"`
	Seq    int64  `json:"seq"`
	Text   string `json:"text"`
	IsEcho bool   `json:"is_echo"`
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
	Messaging_type string    `json:"messaging_type"`
	Recipient      Recipient `json:"recipient"`
	Message        struct {
		Text string `json:"text"`
	} `json:"message"`
}

//Регулярное выражение для запроса данных трек номера Регион курьер Липецк 15 или 17 символов 000020004000085
var ValidRKLIP = regexp.MustCompile(`(?m)^(([0-9]{15})|([0-9]{17}))$`)
var ValidTranslate = regexp.MustCompile(`(?m)(^[a-z-A-Z].*$)`)
var ValidRUSSIANPOST = regexp.MustCompile(`(?m)^(([0-9]{14})|([0-9A-Z]{13}))$`)

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
		log.Println("senderID: %+v \n" + senderID)
		log.Print("%+v\n", event)
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
			SendMessageToBot(senderID, message)
		}
	}
	//fmt.Fprintf(w, "Success")
}

// TODO: Reply message is just sample and made by easy logic, need to enhance the logic.
func getReplyMessage(receivedMessage string) string {
	var message string
	receivedMessage = strings.ToUpper(receivedMessage)
	log.Print(" Received message: " + receivedMessage)
	if ValidRKLIP.MatchString(receivedMessage) == true {
		// Поступил запрос трэк номера РегионКурьер Липецк
		message = req2rkLip(string(receivedMessage))
	} else if ValidRUSSIANPOST.MatchString(receivedMessage) == true {
		// Поступил запрос трэк номера RUSSIANPOST
		message = req2russianpost(string(receivedMessage))
	} else if ValidTranslate.MatchString(receivedMessage) == true {
		// Поступил запрос текста на английском - переведем его.
		message = getTranslate(receivedMessage)
	} else {
		message = `Уточните Штриховой Почтовый Идентификатор, пожалуйста. И повторите запрос.`
	}
	//	if strings.Contains(receivedMessage, "TEST") {
	//		message = "Вы запросили TEST"
	//	} else if strings.Contains(receivedMessage, "TEST1") {
	//		message = "Вы запросили TEST1"
	//	} else {
	//		message = " Ваше сообщение принято"
	//	}
	return message
}

//curl -H "Content-Type: application/json" -X POST "localhost:1337/webhook" -d '{"object": "page", "entry": [{"messaging": [{"message": "TEST_MESSAGE"}]}]}'

// SendMessageToBot sends a message to the Facebook bot
//curl -X POST -H "Content-Type: application/json" -d '{
//"messaging_type": "<MESSAGING_TYPE>",
//	"recipient":{
//	  "id":"<PSID>"
//	},
//	"message":{
//	  "text":"hello, world!"
//	}
//}' "https://graph.facebook.com/v2.6/me/messages?access_token=<PAGE_ACCESS_TOKEN>"

func SendMessageToBot(botID string, rtext string) {

	recipient := new(Recipient)
	log.Printf("botID=%+v\n", botID)
	recipient.ID = botID
	sendMessage := new(SendMessage)
	sendMessage.Messaging_type = "RESPONSE"
	sendMessage.Recipient = *recipient
	sendMessage.Message.Text = rtext
	log.Printf("sendMessage: %+v \n", sendMessage)
	sendMessageBody, err := json.Marshal(sendMessage)
	log.Printf("Marshal %+v\n", string(sendMessageBody))
	if err != nil {
		log.Println("err json.Marshal(sendMessage)")
		log.Print(err)
	}

	
	buffer := new(bytes.Buffer)
	params := url.Values{}
	params.Set("access_token", AccessToken)
//	params.Set("access_token", VerifyToken)
	buffer.WriteString(params.Encode())
	buffer.Write(sendMessageBody)
	
//	req, err := http.NewRequest("POST", FacebookEndPoint, buffer)
//	if err != nil {
//		log.Printf("err http.NewRequest %v %v\n", FacebookEndPoint, sendMessageBody)
//		log.Print(err)
//	}
 //
//	req.Header.Set("content-type", "application/json")
//	client := &http.Client{Timeout: time.Duration(30 * time.Second),
//		Transport: &http.Transport{
//			TLSClientConfig: &tls.Config{
//				InsecureSkipVerify: true,
//			},
//		},
//      }
//
//	log.Printf("req=%+v\n", req)
//	res, err := client.Do(req)

	client := &http.Client{Timeout: time.Duration(30 * time.Second),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
      }
	res, err := client.Post(FacebookEndPoint, `application/json; charset=utf-8`, buffer)

	if err != nil {
//		log.Printf("Ошибка client.Do(req) req=%v\n", req)
		log.Print(err)
	}

	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Ошибка ioutil.ReadAll(res.Body) res.Body= %+v \n ", res.Body)
		log.Print(err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}


// Функция getJoke() string , возвращает строку с шуткой, полученной от сервиса  http://api.icndb.com/jokes/random?limitTo=[nerdy]
func getJoke() string {
	c := http.Client{}
	resp, err := c.Get("http://api.icndb.com/jokes/random?limitTo=[nerdy]")
	if err != nil {
		return "jokes API not responding"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	joke := JokeResponse{}
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "Joke error"
	}
	return joke.Value.Joke
}

// Функция getTranslate(mytext string) string, переводит полученный английский текст на русский или если
// текст отсутствует, получает очередную шутку и возвращает ее перевод на русском
func getTranslate(mytext string) string {
	var sjoke string
	if mytext == "" {
		//Получим очередную шутку на английском
		sjoke = getJoke()
	} else {
		// если поступил англ. текст, принимаем его для перевода
		sjoke = mytext
	}
	c := http.Client{}
	lang := "en-ru"
	// подготовим параметры для POST запроса
	builtParams := url.Values{"key": {Keyyandex}, "lang": {lang}, "text": {sjoke}, "options": {"1"}}
	resp, err := c.PostForm(WebTranslateURL, builtParams)
	if err != nil {
		return "Переводчик yandex API not responding..."
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	tjoke := TranslateJoke{}
	err = json.Unmarshal(body, &tjoke)
	if err != nil {
		serr := fmt.Sprintf("%v", err)
		return "Unmarshal error " + serr
	}
	return strings.Join(tjoke.Text[:], ",")

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/webhook", webhookEndpoint)
	if err := http.ListenAndServe(":"+Port, r); err != nil {
		log.Fatal(err)
	}
}
