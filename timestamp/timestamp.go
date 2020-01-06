package timestamp

import (
	"time"
)

const LAYOUT = "02-01-2006 15:04"

func ParseTimestampMessage(msg string) (time.Time, error) {
	return time.Parse(LAYOUT, msg)
}
