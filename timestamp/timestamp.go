package timestamp

import (
	"time"

	"github.com/araddon/dateparse"
)

func ParseTimestampMessage(msg string) (time.Time, error) {
	return dateparse.ParseAny(msg)
}
