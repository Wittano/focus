package focus

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/wittano/focus/seq"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

type LevelValue int

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
	r     *csv.Reader
	cache posCache
}

func (d *Database) Close() error {
	return d.f.Close()
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
	}

	for {
		lines, err := d.r.Read()
		if err != nil {
			return None, err
		}
		hour := t.Hour()
		isSameDate := bytes.Equal(bytes.Trim([]byte(lines[0]), string([]byte{0x0})), []byte(t.Format(dateFormat)))
		if isSameDate && hour+1 < len(lines) {
			val := lines[hour+1]
			l, err := strconv.Atoi(string(bytes.Trim([]byte(val), string([]byte{0x0}))))
			if err != nil && val != "" {
				log.Println(err)
				break
			} else if val == "" {
				l = 0
			}
			return LevelValue(l), nil
		} else if isSameDate && hour+1 >= len(lines) {
			return None, nil
		}
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
	if prevOffset, ok = d.cache.Position(t); !ok {
		prevOffset, err = d.createEntry(t)
		if err != nil {
			return err
		}
	}

	_, err = d.f.Seek(prevOffset, io.SeekStart)
	if err != nil {
		return err
	}
	rawVal := strconv.Itoa(int(l))

	for {
		prevOffset = d.r.InputOffset()
		lines, err := d.r.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		if lines[0] == t.Format(dateFormat) && t.Hour()+1 < len(lines) {
			var (
				n      int
				offset = int64(seq.SumStringLength(lines[:t.Hour()+1])+t.Hour()+1) + prevOffset
				buf    = make([]byte, 1)
			)

			if n, err = d.f.ReadAt(buf, offset+1); err != nil && n == 1 && buf[0] != ',' {
				_, err = d.f.WriteAt([]byte(","), offset+1)
				if err != nil {
					return err
				}
			}

			_, err = d.f.WriteAt([]byte(rawVal), offset)
			if err == nil {
				go func() {
					if err = d.cache.Update(t, 1); err != nil {
						log.Println(err)
					}
				}()
			}
			return err
		} else if lines[0] == t.Format(dateFormat) && t.Hour()+1 >= len(lines) {
			missingCommas := t.Hour() + 2 - len(lines)
			offset := int64(seq.SumStringLength(lines)) + prevOffset
			putDataCount := 0
			for i := 0; i < missingCommas; i++ {
				var val []byte
				if i == missingCommas-1 {
					val = []byte("," + rawVal)
				} else {
					val = []byte(",")
				}

				putDataCount += len(val)

				_, err = d.f.WriteAt(val, offset+int64(i))
				if err != nil {
					return err
				}
			}

			return d.cache.Update(t, putDataCount)
		}
	}

	return fmt.Errorf("no entry for %s date found", t.Format(dateFormat))
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

	db.r = csv.NewReader(f)
	db.r.FieldsPerRecord = -1
	db.cache, err = newCache(f)
	return
}
