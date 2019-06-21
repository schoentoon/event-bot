package utils

import (
	"encoding/json"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"github.com/getsentry/sentry-go"
)

func Recover(update tgbotapi.Update) {
	if perr := recover(); perr != nil {
		hub := sentry.CurrentHub().Clone()

		data, err := json.Marshal(update)
		if err != nil {
			panic(err)
		}

		hub.Scope().SetExtras(map[string]interface{}{"request":string(data)})

		hub.Recover(perr)
	}
}