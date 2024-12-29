package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"text/template"

	"github.com/gorilla/mux"
)

type WeatherApp struct {
	mu      sync.Mutex
	content []byte
}

func main() {
	app := &WeatherApp{}

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Fatal("WEATHER_API_KEY environment variable not set")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", app.serveHome).Methods("GET", "POST")
	r.HandleFunc("/showResponse", app.showResponse).Methods("GET")

	log.Println("Server started at localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func (app *WeatherApp) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("home.html")
		if err != nil {
			http.Error(w, "Failed to load template", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to execute template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == "POST" {
		place := r.FormValue("place")
		if place == "" {
			http.Error(w, "Place cannot be empty", http.StatusBadRequest)
			return
		}

		apiKey := os.Getenv("WEATHER_API_KEY")
		url := fmt.Sprintf("http://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, place)

		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Failed to fetch weather data", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Weather API returned an error", resp.StatusCode)
			return
		}

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}

		app.mu.Lock()
		app.content = content
		app.mu.Unlock()

		http.Redirect(w, r, "/showResponse", http.StatusSeeOther)
	}
}

func (app *WeatherApp) showResponse(w http.ResponseWriter, r *http.Request) {
	app.mu.Lock()
	content := app.content
	app.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}
