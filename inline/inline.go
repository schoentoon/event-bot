package inline

import (
	"database/sql"

	"gitlab.com/schoentoon/event-bot/database"
	"gitlab.com/schoentoon/event-bot/events"
	"gitlab.com/schoentoon/event-bot/idhash"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func HandleInlineQuery(db *sql.DB, bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) error {
	var idFromQuery int64
	typ, id, err := idhash.Decode(query.Query)
	if err == nil && typ == idhash.Event {
		idFromQuery = id
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DECLARE events_cursor CURSOR FOR
		SELECT id, name, description
		FROM public.events
		WHERE (id = $3 AND publicly_shareable)
		OR ("owner" = $1
			AND (name SIMILAR TO concat('%', $2::text, '%') OR
				description SIMILAR TO concat('%', $2::text, '%') OR
				id = $3))`,
		query.From.ID, query.Query, idFromQuery)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		IsPersonal:    true,
		CacheTime:     0,
		Results:       []interface{}{},
	}

	for {
		var id int64
		var name string
		var description string
		row := tx.QueryRow(`FETCH NEXT FROM events_cursor`)
		err := row.Scan(&id, &name, &description)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return database.TxRollback(tx, err)
		}
		rendered, event, err := events.FormatEvent(tx, id)
		if err != nil {
			return database.TxRollback(tx, err)
		}

		art := tgbotapi.NewInlineQueryResultArticleHTML(idhash.Encode(idhash.Event, id), name, rendered)
		art.Description = description
		art.ReplyMarkup = events.CreateInlineKeyboard(event, id)
		inlineConf.Results = append(inlineConf.Results, art)
	}

	_, err = tx.Exec(`CLOSE events_cursor`)
	if err != nil {
		return database.TxRollback(tx, err)
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		return database.TxRollback(tx, err)
	}
	return tx.Commit()
}

func HandleChoseInlineResult(db *sql.DB, result *tgbotapi.ChosenInlineResult) error {
	typ, id, err := idhash.Decode(result.ResultID)
	if err != nil {
		return err
	}

	switch typ {
	case idhash.Event:
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO public.inline_messages
				(event_id, inline_message_id)
				VALUES
				($1, $2)`,
			id, result.InlineMessageID)
		if err != nil {
			return err
		}

		return tx.Commit()
	}

	return nil
}
