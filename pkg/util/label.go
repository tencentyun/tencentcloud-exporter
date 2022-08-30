package util

import (
	"regexp"
	"unicode"
)

func IsValidTagKey(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return false
		}
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_]+$`).MatchString(str) {
		return false
	}
	return true
}
