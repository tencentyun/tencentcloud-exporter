package util

import "time"

// FormatTime formats time by zone.
func FormatTime(t time.Time, format string) string {
	var local time.Time
	_, offset := t.Zone()
	if offset == 0 {
		local = t.Add(8 * time.Hour)
	} else {
		local = t
	}
	return local.Format(format)
}
