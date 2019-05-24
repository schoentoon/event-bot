package main

import (
	"database/sql"
	"log"
	"strconv"

	"gopkg.in/telegram-bot-api.v4"
)

func handleInlineQuery(db *sql.DB, bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) error {
	rows, err := db.Query(`SELECT id, name, description
		FROM public.events
		WHERE "owner" = $1
		AND insert_state = 'done'
		AND ($2 SIMILAR TO name OR
			 $2 SIMILAR TO description OR
			 id = $2)`,
			query.From.ID)
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
		art := tgbotapi.NewInlineQueryResultArticle(strconv.FormatInt(id, 10), name, description)
		art.Description = description
		inlineConf.Results = append(inlineConf.Results, art)
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		log.Println(err)
	}
	return nil
}