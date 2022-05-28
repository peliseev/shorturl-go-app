package telegram

import (
	"context"
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/peliseev/shorturl-go-app/domain"
	"github.com/peliseev/shorturl-go-app/mongo"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

type Bot struct {
	bot        *tgbot.BotAPI
	urlService domain.ShortURLService
}

func NewBot(apiKey string, db *mongoDriver.Client) *Bot {
	bot, err := tgbot.NewBotAPI(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account: %s", bot.Self.UserName)

	b := Bot{bot: bot}
	b.urlService = mongo.NewShortURLService(db)

	return &b
}

func (b *Bot) Run() {
	uc := tgbot.NewUpdate(0)
	uc.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(uc)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		ctx := context.WithValue(context.Background(), "updateId", update.UpdateID)
		if update.Message != nil {
			log.Printf("New message from %s: %q",
				update.Message.From.UserName, update.Message.Text)
			switch update.Message.Text {
			case "/start":
				b.handleGreeting(&update)
			default:
				b.handleURL(ctx, &update)
			}
		}
	}
}
