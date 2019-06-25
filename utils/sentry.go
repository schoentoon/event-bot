package utils

import (
	"encoding/json"

	"github.com/getsentry/sentry-go"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func Recover(update tgbotapi.Update) {
	if perr := recover(); perr != nil {
		hub := sentry.CurrentHub().Clone()

		data, err := json.Marshal(update)
		if err != nil {
			panic(err)
		}

		hub.Scope().SetExtra("request", string(data))

		hub.Recover(perr)
	}
}
