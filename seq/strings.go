package seq

import (
	"strings"
	"unicode"
)

func SumStringLength(s []string) (l int) {
	for _, v := range s {
		l += len(v)
	}

	return
}

func CleanStringFromUnprintedChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, s)
}
