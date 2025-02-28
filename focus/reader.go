package focus

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type LevelValue int

func (l LevelValue) String() string {
	var s string
	switch l {
	case Flow:
		s = "Flow"
	case VeryLow:
		s = "Very low"
	case Medium:
		s = "Medium"
	case High:
		s = "High"
	case Low:
		s = "Low"
	case None:
		s = "None"
	default:
		s = "Unknown"
	}

	return s
}

const (
	None LevelValue = iota
	VeryLow
	Low
	Medium
	High
	Flow
)

const (
	database   = "focus.csv"
	dateFormat = "02.01.2006"
)

type Database struct {
	f   *os.File
	csv *csv.Reader
}

func (f *Database) Close() error {
	return f.f.Close()
}

var ErrNotFound = errors.New("no entry found")

func (f *Database) Level(t time.Time) (LevelValue, error) {
	if t.Compare(time.Now()) > 0 {
		return None, errors.New("date entry cannot be in the future")
	}

	date := t.Format(dateFormat)
	defer f.f.Seek(0, io.SeekStart)

	for {
		lines, err := f.csv.Read()
		if err != nil {
			return None, err
		}
		hour := t.Hour()
		if lines[0] == date && hour+1 < len(lines) {
			val := lines[hour+1]
			l, err := strconv.Atoi(val)
			if err != nil {
				log.Println(err)
				break
			}
			return LevelValue(l), nil
		} else if lines[0] == date && hour+1 >= len(lines) {
			return None, nil
		}
	}

	return None, ErrNotFound
}

func New(csvPath string) (db *Database, err error) {
	p := csvPath
	flag := os.O_RDONLY
	if p == "" {
		flag |= os.O_CREATE
		p = database
	}

	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			err = f.Close()
		}
	}()
	db = new(Database)
	db.f = f

	stat, err := f.Stat()
	if err != nil {
		return
	}

	if stat.Size() == 0 || !hasTodayEntry(f) {
		err = createTodayEntry(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	db.csv = csv.NewReader(f)
	db.csv.FieldsPerRecord = -1
	return
}

func hasTodayEntry(r io.ReadSeeker) bool {
	if _, err := r.Seek(-1, io.SeekEnd); err != nil {
		log.Println(err)
		return false
	}
	defer r.Seek(0, io.SeekStart)

	buf := make([]byte, 64)
	var (
		err error
		i   int64
	)
	for i = -2; ; i-- {
		_, err = r.Read(buf)
		if err != nil {
			break
		}
		if buf[0] == '\n' {
			_, err = r.Seek(i+1, io.SeekEnd)
			for j := 1; j < len(buf); j++ {
				if buf[j] == ',' {
					buf = buf[1:j]
					break
				}
			}
			break
		}
		_, err = r.Seek(i, io.SeekEnd)
		if err != nil {
			break
		}
	}

	if err != nil {
		log.Println(err)
		return false
	}

	date, err := time.Parse(dateFormat, string(buf))
	if err != nil {
		log.Println(err)
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return date.Compare(today) >= 0
}

func createTodayEntry(w io.WriteSeeker) error {
	_, err := w.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	defer w.Seek(0, io.SeekStart)

	_, err = w.Write([]byte(fmt.Sprintf("\n%s,", time.Now().Format(dateFormat))))
	return err
}
