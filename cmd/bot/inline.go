package main

import (
	"database/sql"
	"log"
	"fmt"
	"strconv"

	"gitlab.schoentoon.com/schoentoon/event-bot/utils"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func handleInlineQuery(db *sql.DB, bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) error {
	idFromQuery, err := strconv.ParseInt(query.Query, 10, 64)
	if err != nil {
		idFromQuery = -1
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
		art := tgbotapi.NewInlineQueryResultArticleHTML(fmt.Sprintf("%d", id), name, "<b>Shared text</b>: "+description)
		art.Description = description
		art.ReplyMarkup = utils.CreateInlineKeyboard(id)
		inlineConf.Results = append(inlineConf.Results, art)
	}

	if _, err := bot.AnswerInlineQuery(inlineConf); err != nil {
		log.Println(err)
	}
	return nil
}
