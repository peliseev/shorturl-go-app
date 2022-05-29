package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/peliseev/shorturl-go-app/mongo"
	"github.com/peliseev/shorturl-go-app/server"
	"github.com/peliseev/shorturl-go-app/telegram"
)

type config struct {
	port, mongoUrl, tgAPIkey, urlPrefix string
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
	service := mongo.NewShortURLService(db)
	bot := telegram.NewBot(cfg.urlPrefix, cfg.tgAPIkey, db, service)
	server := server.NewServer(db, service)
	go bot.Run()
	log.Fatal(server.Run(cfg.port))
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

	urlPrefix, ok := os.LookupEnv("URL_PREFIX")
	if !ok {
		urlPrefix = "http://localhost:8080/"
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = ":8080"
	}

	return config{
		mongoUrl:  mongoUrl,
		tgAPIkey:  tgAPIkey,
		urlPrefix: urlPrefix,
		port:      port,
	}
}
