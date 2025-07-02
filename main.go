package main

import (
	"log"
	"my-telegram-bot/config"
	"my-telegram-bot/internal/bot"
	"my-telegram-bot/internal/database"
)

func main() {
	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	bot, err := bot.NewBot(cfg.BotToken, db)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	bot.Start()
}
