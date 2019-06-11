package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"sync"

	"gitlab.schoentoon.com/schoentoon/event-bot/database"
	"gitlab.schoentoon.com/schoentoon/event-bot/events"
	"gitlab.schoentoon.com/schoentoon/event-bot/idhash"
	"gitlab.schoentoon.com/schoentoon/event-bot/templates"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func main() {
	var cfgfile = flag.String("config", "config.yml", "Config file location")
	flag.Parse()

	cfg, err := ReadConfig(*cfgfile)
	if err != nil {
		panic(err)
	}

	err = idhash.InitHasher(cfg.IDHash.Salt, cfg.IDHash.MinLength)
	if err != nil {
		panic(err)
	}

	err = templates.Load(cfg.Templates)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", cfg.Postgres.Addr)
	if err != nil {
		panic(err)
	}
	cfg.ApplyDatabase(db)
	err = database.UpgradeDatabase(db)
	if err != nil {
		panic(err)
	}

	if cfg.Prometheus.ListenAddr != "" {
		go func(db *sql.DB) {
			dbCollector := database.NewCollector(db)
			prometheus.MustRegister(dbCollector)
			prometheus.MustRegister(prometheus.NewBuildInfoCollector())
			http.Handle("/metrics", promhttp.Handler())
			log.Fatal(http.ListenAndServe(cfg.Prometheus.ListenAddr, nil))
		}(db)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		panic(err)
	}
	if cfg.Telegram.Debug {
		log.Println("Enabling Telegram debug mode")
		bot.Debug = true
	}

	go events.UpdateLoop(db, bot, cfg.EventRefreshInterval)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic(err)
	}

	log.Printf("Starting %d workers", cfg.Workers)

	wg := &sync.WaitGroup{}
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go worker(wg, db, bot, updates)
	}

	wg.Wait()
}
