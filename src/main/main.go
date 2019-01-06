package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/gorilla/mux"
)

// Keytg ключ бота Facebook
var Keyfb string
// Keytg ключ сервиса на yandex
var Keyyandex string
var Port string
//  WebhookURL url сервера бота
const WebhookURL = "https://app2-test48.herokuapp.com/"

// WebTranslateURL url сервиса переводчика на русский с английского
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"
//Структура данных ответа от https://translate.yandex.net/api/v1.5/tr.json/translate
// Пример на запрос Hellow Jack, ответ: {"code":200,"lang":"en-ru","text":["\"Хеллоу Джек\""]}
type TranslateJoke struct {
	CODE uint32   `json: "code"`
	Lang string   `json: "lang"`
	Text []string `json: "text"`
}

//Проинициализируем ключи Keyfb Keyyandex
func init() {
	Port = os.Getenv("PORT")
	Keyfb = os.Getenv("KEYFB")
	Keyyandex = os.Getenv("KEYYANDEX")

}

func HomeEndpoint(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello :)")
}


func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeEndpoint)
	if err := http.ListenAndServe(":"+Port, r); err != nil {
		log.Fatal(err)
	}
}