package util

func ToUnderlineLower(s string) string {
	var interval byte = 'a' - 'A'
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += interval
			if i != 0 {
				b = append(b, '_')
			}
		}
		b = append(b, c)
	}
	return string(b)
}
