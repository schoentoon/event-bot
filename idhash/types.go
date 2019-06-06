package idhash

//go:generate stringer -type=HashType
type HashType int64

const (
	Invalid HashType = iota
	Event
	VoteYes
	VoteMaybe
	VoteNo
)
