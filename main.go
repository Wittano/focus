package main

import (
	"flag"
	"fmt"
	"github.com/wittano/focus/focus"
	"log"
	"time"
)

var path = flag.String("path", "", "path to focus-date.csv file")

func init() {
	flag.Parse()
}

func main() {
	if path == nil {
		path = new(string)
	}

	db, err := focus.New(*path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println(db.Level(time.Now()))
}
