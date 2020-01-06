package timestamp

import (
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	// Thu Jan 9th 2020 19:30
	date, err := ParseTimestampMessage("09-01-2020 19:30")
	if err != nil {
		t.Fatal(err)
	}
	expect := time.Date(2020, time.January, 9, 19, 30, 0, 0, time.UTC)
	if date.Equal(expect) == false {
		t.Fatalf("Expected %s, got %s", expect, date)
	}
}
