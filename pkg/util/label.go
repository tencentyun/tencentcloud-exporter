package util

import (
	"regexp"
	"unicode"
)

// IsValidTagKey ref: https://cloud.tencent.com/document/product/1416/73370#.E8.87.AA.E5.AE.9A.E4.B9.89.E4.B8.8A.E6.8A.A5.E9.99.90.E5.88.B6
func IsValidTagKey(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return false
		}
	}
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(str) || len(str) > 1024 {
		return false
	}
	return true
}
