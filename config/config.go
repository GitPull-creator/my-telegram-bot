package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	BotToken string
	DBPath   string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
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
