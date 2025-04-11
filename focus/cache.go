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
	"unicode"
)

type posCache struct {
	mutex     sync.Mutex
	positions map[string]int64
}

func (p *posCache) Position(t time.Time) (int64, bool) {
	p.mutex.Lock()
	pos, ok := p.positions[t.Format(dateFormat)]
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

	p.mutex.Lock()
	p.positions[t.Format(dateFormat)] = pos
	p.mutex.Unlock()

	return nil
}

func newCache(f *os.File) (cache posCache, err error) {
	if f == nil {
		panic("focus: file pointer is nil")
	}
	f.Seek(0, io.SeekStart)
	defer f.Seek(0, io.SeekStart)

	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return
	}

	cache = posCache{positions: make(map[string]int64)}

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		txt := s.Text()
		split := strings.Split(txt, ",")
		if len(split) == 0 {
			continue
		}

		var rawTime strings.Builder
		for _, r := range split[0] {
			if unicode.IsGraphic(r) {
				rawTime.WriteRune(r)
			}
		}

		var t time.Time
		t, err = time.Parse(dateFormat, rawTime.String())
		if err != nil {
			return
		}
		if err = cache.Put(t, offset, 0); err != nil && !errors.Is(err, ErrInvalidTime) {
			return
		} else if errors.Is(err, ErrInvalidTime) {
			err = nil
			break
		}

		for _, s := range split {
			offset += int64(len(s)) + 1
		}
	}

	return
}
