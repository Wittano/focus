package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	database   = "focus.csv"
	dateFormat = "2006-01-02"
)

func main() {
	f, err := os.OpenFile(database, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	stat, err := f.Stat()
	if err != nil {
		log.Fatal(err)
		return
	}

	if stat.Size() == 0 || !checkTodayEntry(f) {
		err = createTodayEntry(f)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkTodayEntry(r io.ReadSeeker) bool {
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

	return compareDate(date, time.Now())
}

func compareDate(l time.Time, r time.Time) bool {
	lyyyy, lmm, ldd := l.Date()
	ryyyy, rmm, rdd := r.Date()

	return lyyyy == ryyyy && lmm == rmm && ldd == rdd
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
