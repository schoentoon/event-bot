package inline

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleInlineQuery(db *sql.DB, bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) error {
	var idFromQuery int64
	var err error
	split := strings.Split(query.Query, "/")
	if len(split) == 2 {
		idFromQuery, err = strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			idFromQuery = -1
		}
	}

	rows, err := db.Query(`SELECT id, name, description
		FROM public.events
		WHERE "owner" = $1
		AND (name SIMILAR TO concat('%', $2::text, '%') OR
			 description SIMILAR TO concat('%', $2::text, '%') OR
			 id = $3)`,
		query.From.ID, query.Query, idFromQuery)
	if err != nil {
		return err
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{},
	}

	for rows.Next() {
		var id int64
		var name string
		var description string
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			return err
		}
		rendered, err := utils.FormatEvent(db, id)
		if err != nil {
			return err
		}

		art := tgbotapi.NewInlineQueryResultArticleHTML(fmt.Sprintf("event/%d", id), name, rendered)
		art.Description = description
		art.ReplyMarkup = utils.CreateInlineKeyboard(id)
		inlineConf.Results = append(inlineConf.Results, art)
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		return err
	}
	return nil
}

func HandleChoseInlineResult(db *sql.DB, result *tgbotapi.ChosenInlineResult) error {
	split := strings.Split(result.ResultID, "/")
	if len(split) < 2 {
		return errors.New("Split is less than 2, ignoring")
	}

	if split[0] == "event" {
		eventID, err := strconv.ParseInt(split[1], 10, 64)
		if err != nil {
			return err
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO public.inline_messages
				(event_id, inline_message_id)
				VALUES
				($1, $2)`,
			eventID, result.InlineMessageID)
		if err != nil {
			return err
		}

		return tx.Commit()
	}

	return nil
}
