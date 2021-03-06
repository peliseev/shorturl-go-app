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
	port, mongoURL, tgAPIkey, urlPrefix string
}

func init() {
	err := godotenv.Load("/etc/.env")
	if err != nil {
		log.Print("Error loading .env file")
	}
}

func main() {
	cfg := envConfig()
	db := mongo.Open(cfg.mongoURL)
	service := mongo.NewShortURLService(db)
	bot := telegram.NewBot(cfg.urlPrefix, cfg.tgAPIkey, service)
	srv := server.NewServer(db, service)
	go bot.Run()
	log.Fatal(srv.Run(cfg.port))
}

func envConfig() config {
	mongoURL, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		mongoURL = "mongodb://localhost:27017/"
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
		mongoURL:  mongoURL,
		tgAPIkey:  tgAPIkey,
		urlPrefix: urlPrefix,
		port:      port,
	}
}
