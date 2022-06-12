package telegram

import (
	"log"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/peliseev/shorturl-go-app/domain"
	"github.com/peliseev/shorturl-go-app/mongo"
)

type Bot struct {
	URLPrefix  string
	Bot        IBotAPI
	URLService domain.ShortURLService
}

type IBotAPI interface {
	Send(c tgbot.Chattable) (tgbot.Message, error)
	GetUpdatesChan(config tgbot.UpdateConfig) (tgbot.UpdatesChannel, error)
}

func NewBot(urlPrefix, apiKey string, sus *mongo.ShortURLService) *Bot {
	bot, err := tgbot.NewBotAPI(apiKey)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account: %s", bot.Self.UserName)

	return &Bot{
		URLPrefix:  urlPrefix,
		Bot:        bot,
		URLService: sus,
	}
}

func (b *Bot) Run() {
	uc := tgbot.NewUpdate(0)
	uc.Timeout = 60

	updates, err := b.Bot.GetUpdatesChan(uc)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message != nil {
			log.Printf("New message from %s: %q",
				update.Message.From.UserName, update.Message.Text)
			switch update.Message.Text {
			case "/start":
				b.HandleGreeting(&update)
			default:
				b.HandleURL(&update)
			}
		}
	}
}
