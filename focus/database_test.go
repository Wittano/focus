package focus

import (
	"os"
	"testing"
	"time"
)

var data = map[string]LevelValue{
	"28.02.2025":                  High,
	"21.01.2025":                  None,
	"21.12.2024":                  VeryLow,
	"11.02.2001":                  Low,
	time.Now().Format(dateFormat): Medium,
	"29.02.2024":                  Flow,
}

func TestDatabase_Put(t *testing.T) {
	csv, err := os.CreateTemp(t.TempDir(), database)
	if err != nil {
		t.Fatal(err)
	}
	csv.Close()

	db, err := New(csv.Name())
	if err != nil {
		t.Fatal(err)
	}

	for date, exp := range data {
		ti, err := time.Parse(dateFormat, date)
		if err != nil {
			t.Fatal(err)
		}

		t.Run(date, func(t *testing.T) {
			if err = db.Put(ti, exp); err != nil {
				t.Fatal(err)
				return
			}

			got, err := db.Level(ti)
			if err != nil {
				t.Fatal(err)
				return
			}

			if exp != got {
				t.Errorf("invalid focus level: got %d, want %d", got, exp)
			}
		})
	}
}
