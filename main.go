package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/peliseev/shorturl-go-app/mongo"
	"github.com/peliseev/shorturl-go-app/telegram"
)

type config struct {
	mongoUrl, tgAPIkey string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}
}

func main() {
	cfg := envConfig()
	db := mongo.Open(cfg.mongoUrl)
	bot := telegram.NewBot(cfg.tgAPIkey, db)
	bot.Run()
}

func envConfig() config {
	mongoUrl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		mongoUrl = "mongodb://localhost:27017/"
	}

	tgAPIkey, ok := os.LookupEnv("TELEGRAM_BOT_API_KEY")
	if !ok {
		log.Fatal("TELEGRAM_BOT_API_KEY should be provided")
	}

	return config{mongoUrl: mongoUrl, tgAPIkey: tgAPIkey}
}
