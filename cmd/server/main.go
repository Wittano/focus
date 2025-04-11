package main

import (
	"errors"
	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/wittano/focus/components"
	"github.com/wittano/focus/focus"
	"log"
	"net/http"
	"strconv"
	"time"
)

const port = 3000

func handleTemplate(c *templ.ComponentHandler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c.ServeHTTP(w, r)
	}
}

var emptyData = make([]focus.LevelValue, 24)

func main() {
	db, err := focus.New("./focus-plot.csv")
	if err != nil {
		log.Fatal(err)
	}

	t := time.Now()
	data, err := db.Levels(t)
	if err != nil && !errors.Is(err, focus.ErrNotFound) {
		log.Fatal(err)
	}
	if data == nil {
		data = emptyData
	}

	r := chi.NewRouter()
	r.Get("/", handleTemplate(templ.Handler(components.Home(t, data))))
	r.Get("/data", func(w http.ResponseWriter, r *http.Request) {
		rawData := r.FormValue("date")
		if rawData == "" {
			log.Println("Missing required parameter 'data'")
		}

		t, err := time.Parse(time.DateOnly, rawData)
		if err != nil && rawData != "" {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if rawData == "" {
			t = time.Now()
		}

		if t.Compare(time.Now()) > 0 {
			log.Println("/data: time 'data' cannot be from future")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		data, err := db.Levels(t)
		if err != nil && !errors.Is(err, focus.ErrNotFound) {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if errors.Is(err, focus.ErrNotFound) {
			data = emptyData
		}

		templ.Handler(components.FocusData(t, data)).ServeHTTP(w, r)
	})

	log.Println("Listening on " + strconv.Itoa(port))
	if err = http.ListenAndServe(":"+strconv.Itoa(port), r); err != nil {
		log.Fatal(err)
	}
}
