package telegram

import (
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/peliseev/shorturl-go-app/domain"
	"github.com/peliseev/shorturl-go-app/mongo"
	driver "go.mongodb.org/mongo-driver/mongo"
)

type Bot struct {
	urlPrefix  string
	bot        *tgbot.BotAPI
	urlService domain.ShortURLService
}

func NewBot(urlPrefix, apiKey string, db *driver.Client, sus *mongo.ShortURLService) *Bot {
	bot, err := tgbot.NewBotAPI(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account: %s", bot.Self.UserName)

	return &Bot{
		urlPrefix:  urlPrefix,
		bot:        bot,
		urlService: sus,
	}
}

func (b *Bot) Run() {
	uc := tgbot.NewUpdate(0)
	uc.Timeout = 60

	updates, err := b.bot.GetUpdatesChan(uc)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			log.Printf("New message from %s: %q",
				update.Message.From.UserName, update.Message.Text)
			switch update.Message.Text {
			case "/start":
				b.handleGreeting(&update)
			default:
				b.handleURL(&update)
			}
		}
	}
}
