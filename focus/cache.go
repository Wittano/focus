package focus

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type posCache struct {
	mutex     sync.Mutex
	positions map[time.Time]int64
}

func (p *posCache) Position(t time.Time) (int64, bool) {
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	p.mutex.Lock()
	pos, ok := p.positions[t]
	p.mutex.Unlock()
	return pos, ok
}

func (p *posCache) Update(t time.Time, dataLen int) error {
	currentPos, ok := p.Position(t)
	if !ok {
		return fmt.Errorf("focus: no entry for %s", t)
	}

	for k, pos := range p.positions {
		if currentPos >= pos {
			continue
		}

		p.mutex.Lock()
		p.positions[k] += int64(dataLen)
		p.mutex.Unlock()
	}

	return nil
}

var ErrInvalidTime = errors.New("cache: time cannot be from future or zero")

func (p *posCache) Put(t time.Time, pos int64, dataLen int) (err error) {
	if t.IsZero() || t.After(time.Now()) {
		return ErrInvalidTime
	}
	if _, ok := p.Position(t); ok {
		return p.Update(t, dataLen)
	}

	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	p.mutex.Lock()
	p.positions[t] = pos
	p.mutex.Unlock()

	return nil
}

func newCache(f *os.File) (cache posCache, err error) {
	if f == nil {
		panic("focus: file pointer is nil")
	}
	defer f.Seek(0, io.SeekStart)

	buf := make([]byte, 64)
	s := bufio.NewScanner(f)
	s.Buffer(buf, 64)

	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return
	}

	cache = posCache{positions: make(map[time.Time]int64)}
	for s.Scan() {
		txt := s.Text()
		split := strings.Split(txt, ",")
		if len(split) == 0 {
			continue
		}

		rawTime := split[0]
		if rawTime == "" {
			continue
		}

		var t time.Time
		t, err = time.Parse(dateFormat, rawTime)
		if err != nil {
			return
		}
		cache.Put(t, offset, 0)

		offset, err = f.Seek(0, io.SeekCurrent)
		if err != nil {
			return
		}
	}

	return
}
