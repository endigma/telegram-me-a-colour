package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type colorInfo struct {
	name  string
	color string
}

type configType struct {
	botKey    string
	channelId string
}

var Config configType

func main() {
	Config = getConfig()
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(httprate.Limit(
		3,
		1*time.Minute,
		httprate.WithLimitHandler(handleTooManyRequests),
		httprate.WithKeyFuncs(httprate.KeyByIP)))
	router.Post("/", handlePost)

	http.ListenAndServe(":3333", router)
}

func getConfig() configType {
	config := configType{}

	config.botKey = os.Getenv("BOT_KEY")
	checkVar(config.botKey, "BOT_KEY")
	config.channelId = os.Getenv("BOT_CHANNEL_ID")
	checkVar(config.channelId, "BOT_CHANNEL_ID")

	return config
}

func checkVar(val string, name string) {
	if len(val) == 0 {
		log.Fatalf("set environment variable %s!\n", name)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	nameStr := r.FormValue("name")
	namePattern := "^[\\w\\d ]+$"
	colStr := r.FormValue("colour")
	colPattern := "^#([\\dabcdef]{6})$"

	nameMatch, err := regexp.Match(namePattern, []byte(nameStr))
	CheckErr(err)
	colMatch, err := regexp.Match(colPattern, []byte(colStr))
	CheckErr(err)

	tpl, err := template.ParseFiles("template.html")
	CheckErr(err)

	if !nameMatch {
		nameStr = "unknown"
	}

	if !colMatch {
		w.WriteHeader(400)
		tpl.Execute(w, map[string]interface{}{
			"success":          false,
			"requestStatusStr": "the data you sent was invalid!",
		})
	} else {
		w.WriteHeader(200)
		go handleSuccess(nameStr, colStr)
		tpl.Execute(w, map[string]interface{}{
			"success":          true,
			"requestStatusStr": "colour submitted!",
		})
	}
}

func handleTooManyRequests(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("template.html")
	CheckErr(err)
	w.WriteHeader(429)
	tpl.Execute(w, map[string]interface{}{
		"success":          false,
		"requestStatusStr": "you are sending requests too quickly!",
	})

}

func handleSuccess(name string, colour string) {
	size := 256

	botUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendPhoto?chat_id=%s&photo=https://www.singlecolorimage.com/get/%s/%dx%d&caption=%s sent colour %s", Config.botKey, Config.channelId, colour[1:], size, size, name, url.QueryEscape(colour))
	resp, err := http.Get(botUrl)
	CheckErr(err)
	if resp.StatusCode != http.StatusOK {
		log.Printf("err: failed to submit %s from %s\n", colour, name)
	}
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
