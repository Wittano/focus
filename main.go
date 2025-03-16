package main

import (
	"flag"
	"fmt"
	"github.com/wittano/focus/focus"
	"log"
	"time"
)

const timeFormat = "02.01.2006 15:04:05"

var (
	path       = flag.String("path", "", "path to focus-date.csv file")
	focusLevel = flag.Uint("focusLevel", 0, "level of focus")
	date       = flag.String("level", "", "level of focus")
	find       = flag.String("find", "", "get focus date by date in format: "+timeFormat)
)

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

	if find != nil && *find != "" {
		var l focus.LevelValue
		l, err = findLevel(db)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Your focues level at %s was %s(%d)", *find, l, l)
	} else {
		if err = updateDb(db); err != nil {
			log.Fatal(err)
		}

		l := focus.LevelValue(*focusLevel)

		if date == nil {
			date = new(string)
			*date = time.Now().Format(timeFormat)
		} else if *date == "" {
			*date = time.Now().Format(timeFormat)
		}
		fmt.Printf("Your focues level at %s was set on %s(%d)", *date, l, l)
	}
}

func findLevel(db *focus.Database) (focus.LevelValue, error) {
	if db == nil {
		panic("database is nil")
	}

	t, err := time.Parse(timeFormat, *find)
	if err != nil {
		return 0, err
	}

	return db.Level(t)
}

func updateDb(db *focus.Database) (err error) {
	var (
		focusVal focus.LevelValue
		t        time.Time
	)
	if focusLevel != nil {
		focusVal = focus.LevelValue(*focusLevel)
	}

	if date == nil || *date == "" {
		t = time.Now()
		fmt.Printf("No focus date. A new focus date will be set on %s\n", t.Format(timeFormat))
	} else {
		t, err = time.Parse(timeFormat, *date)
		if err != nil {
			return
		}
	}

	return db.Put(t, focusVal)
}
