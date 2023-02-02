package util

import "time"

// FormatTime formats time by zone.
func FormatTime(t time.Time, format string) string {
	var local time.Time
	_, offset := t.Zone()
	local = t.Add(time.Duration(8*3600-offset) * time.Second)
	return local.Format(format)
}
