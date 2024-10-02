package time_string

import (
	"strings"
	"time"
)

const DefaultEmptyDurationString = "0s"

func ShortDur(d time.Duration) string {
	s := d.String()

	if strings.Contains(s, "m0s") {
		s = strings.Replace(s, "0s", "", 1)
	}
	if strings.Contains(s, "h0m") {
		s = strings.Replace(s, "0m", "", 1)
	}

	s = strings.TrimPrefix(s, "0h")
	s = strings.TrimPrefix(s, "0m")
	s = strings.TrimPrefix(s, "0s")

	if s == "" {
		return DefaultEmptyDurationString
	}

	return s
}
