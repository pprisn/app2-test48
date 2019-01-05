package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/paked/messenger"
)

// WebTranslateURL url сервиса переводчика на русский с английского
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"
//Структура данных ответа от https://translate.yandex.net/api/v1.5/tr.json/translate
// Пример на запрос Hellow Jack, ответ: {"code":200,"lang":"en-ru","text":["\"Хеллоу Джек\""]}
type TranslateJoke struct {
	CODE uint32   `json: "code"`
	Lang string   `json: "lang"`
	Text []string `json: "text"`
}

// Keytg ключ бота Telegramm
var Keytg string

// Keytg ключ сервиса на yandex
var Keyyandex string

//Проинициализируем ключи Keytg Keyyandex
func init() {
	Keytg = os.Getenv("KEYTG")
	Keyyandex = os.Getenv("KEYYANDEX")
}

//При старте приложения, оно скажет телеграму ходить с обновлениями по этому URL

//  WebhookURL url сервера бота
const WebhookURL = "https://app2-test48.herokuapp.com/"

// WebTranslateURL url сервиса переводчика на русский с английского
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"

var (

	verifyToken = flag.String("verify-token", "mad-skrilla", "The token used to verify facebook (required)")
	verify      = flag.Bool("should-verify", false, "Whether or not the app should verify itself")
	pageToken   = flag.String("page-token", "not skrilla", "The token that is used to verify the page on facebook")
	appSecret   = flag.String("app-secret", "", "The app secret from the facebook developer portal (required)")
	host        = flag.String("host", "localhost", "The host used to serve the messenger bot")
	port        = flag.Int("port", 8080, "The port used to serve the messenger bot")
)

func main() {
	flag.Parse()

	if *verifyToken == "" || *appSecret == "" || *pageToken == "" {
		fmt.Println("missing arguments")
		fmt.Println()
		flag.Usage()

		os.Exit(-1)
	}

	// Create a new messenger client
	client := messenger.New(messenger.Options{
		Verify:      *verify,
		AppSecret:   *appSecret,
		VerifyToken: *verifyToken,
		Token:       *pageToken,
	})

	// Setup a handler to be triggered when a message is received
	client.HandleMessage(func(m messenger.Message, r *messenger.Response) {
		fmt.Printf("%v (Sent, %v)\n", m.Text, m.Time.Format(time.UnixDate))

		p, err := client.ProfileByID(m.Sender.ID, []string{"name", "first_name", "last_name", "profile_pic"})
		if err != nil {
			fmt.Println("Something went wrong!", err)
		}

		r.Text(fmt.Sprintf("Hello, %v!", p.FirstName), messenger.ResponseType)
	})

	// Setup a handler to be triggered when a message is delivered
	client.HandleDelivery(func(d messenger.Delivery, r *messenger.Response) {
		fmt.Println("Delivered at:", d.Watermark().Format(time.UnixDate))
	})

	// Setup a handler to be triggered when a message is read
	client.HandleRead(func(m messenger.Read, r *messenger.Response) {
		fmt.Println("Read at:", m.Watermark().Format(time.UnixDate))
	})

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Println("Serving messenger bot on", addr)
	log.Fatal(http.ListenAndServe(addr, client.Handler()))
}