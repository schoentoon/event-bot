package idhash

//go:generate stringer -type=HashType
type HashType int64

const (
	Invalid HashType = iota
	Event

	// these are for changing your status on an existing event
	VoteYes
	VoteMaybe
	VoteNo

	// entering the settings menu
	MainMenu
	Settings

	// the change answers submenu
	SettingChangeAnswers
	ChangeAnswerYesNoMaybe
	ChangeAnswerYesMaybe
	ChangeAnswerYesNo
	ChangeAnswerYes
)
