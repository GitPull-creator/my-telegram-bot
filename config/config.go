package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
	DBPath   string
}

func LoadConfig() Config {
	// Попытаемся загрузить .env файл, но не будем падать если его нет (production)
	_ = godotenv.Load()
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN is not set")
	}
	dbPath := os.Getenv("DB_PATH")
	if len(dbPath) == 0 {
		dbPath = "./bot.db"
	}
	return Config{
		BotToken: token,
		DBPath:   dbPath,
	}
}
