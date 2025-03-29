package focus

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type LevelValue byte

const (
	None LevelValue = iota
	VeryLow
	Low
	Medium
	High
	Flow
)

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
	database   = "focus.csv"
	dateFormat = "02.01.2006"
)

var (
	ErrNotFound   = errors.New("database: no entry found")
	ErrDateFuture = errors.New("database: date entry cannot be in the future")
)

type Database struct {
	f     *os.File
	buf   []byte
	cache posCache
}

func (d *Database) Close() error {
	return d.f.Close()
}

func (d *Database) Levels(t time.Time) ([]LevelValue, error) {
	if pos, ok := d.cache.Position(t); ok {
		_, err := d.f.Seek(pos, io.SeekStart)
		if err != nil {
			return nil, err
		}

		defer d.f.Seek(0, io.SeekStart)
	}

	return d.readLine()
}

func (d *Database) readLine() ([]LevelValue, error) {
	if _, err := d.f.Read(d.buf); err != nil {
		return nil, err
	}

	levels := make([]LevelValue, 24)
	s := strings.Split(string(d.buf), ",")
	for i := 1; i < len(s); i++ {
		l, err := strconv.Atoi(s[i])
		if err != nil {
			return nil, err
		}

		levels[i-1] = LevelValue(l)
	}

	return levels, nil
}

func (d *Database) Level(t time.Time) (LevelValue, error) {
	if t.Compare(time.Now()) > 0 {
		return None, ErrDateFuture
	}

	defer d.f.Seek(0, io.SeekStart)

	if pos, ok := d.cache.Position(t); ok {
		_, err := d.f.Seek(pos, io.SeekStart)
		if err != nil {
			return None, err
		}

		lines, err := d.readLine()
		if err != nil {
			return None, err
		}

		return lines[t.Hour()+1], nil
	}

	return None, ErrNotFound
}

func (d *Database) Put(t time.Time, l LevelValue) error {
	if t.Compare(time.Now()) > 0 {
		return ErrDateFuture
	}
	defer d.f.Seek(0, io.SeekStart)
	var (
		ok         bool
		prevOffset int64
		err        error
	)
	if prevOffset, ok = d.cache.Position(t); ok {
		_, err = d.f.Seek(prevOffset, io.SeekStart)
		if err != nil {
			return err
		}

		buf := bufio.NewReader(d.f)
		line, err := buf.ReadBytes(byte('\n'))
		if err != nil {
			return err
		}

		var (
			commaCount       = 0
			addOffset  int64 = 0
		)
		for _, b := range line {
			if commaCount > t.Hour() {
				break
			}

			if b == byte(',') {
				commaCount++
			}

			addOffset++
		}

		str := []byte(strconv.Itoa(int(l)))
		_, err = d.f.WriteAt(str, prevOffset+addOffset)
		return err
	}

	_, err = d.createEntry(t)
	if err != nil {
		return err
	}

	_, err = d.f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	rawVal := strconv.Itoa(int(l))

	data := strings.Repeat(",", t.Hour()) + rawVal
	_, err = d.f.WriteString(data)
	return err
}

func (d *Database) createEntry(t time.Time) (pos int64, err error) {
	format := t.Format(dateFormat)
	rawDate := []byte("\n" + format)
	if _, ok := d.cache.Position(t); ok {
		return 0, fmt.Errorf("focus: entry on time %s existed", format)
	}

	pos, err = d.f.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	defer d.f.Seek(0, io.SeekStart)
	if pos == 0 {
		rawDate = rawDate[1:]
	}

	_, err = d.f.Write(rawDate)
	if err != nil {
		return
	}
	return pos, d.cache.Put(t, pos, 0)
}

func New(csvPath string) (db *Database, err error) {
	p := csvPath
	flag := os.O_RDWR
	if p == "" {
		flag |= os.O_CREATE
		p = database
	}

	f, err := os.OpenFile(p, flag, 0644)
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
	db.buf = make([]byte, len(dateFormat)+24+24)

	db.cache, err = newCache(f)
	if err != nil {
		err = errors.Join(err, db.f.Close())
	}
	return db, err
}
