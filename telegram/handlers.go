package telegram

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"golang.org/x/net/idna"
	"log"
	"net"
	"net/url"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/peliseev/shorturl-go-app/domain"
)

const (
	GreetingFormat = `Привет %s! 

Этот бот умеет работать только с ссылками. 
Прежде чем прислать ссылку, убедитесь в том, что:
	- домен указан верно; 
	- сайт онлайн;
	- перед ссылкой нет лишних символов.`

	ErrorMsg = `Этот бот умеет работать только с ссылками.
Прежде чем прислать ссылку, убедитесь в том, что:
 - домен указан верно; 
 - сайт онлайн;
 - перед ссылкой нет лишних символов.`
)

func (b *Bot) HandleGreeting(update *tgbot.Update) {
	response(fmt.Sprintf(GreetingFormat, update.Message.From.UserName), b, update)
}

func (b *Bot) HandleURL(update *tgbot.Update) {
	pURL := update.Message.Text
	ok := validateURL(pURL)

	if !ok {
		response(ErrorMsg, b, update)
		return
	}

	user := update.Message.From.UserName

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortURL, err := b.URLService.CreateShortURL(ctx, &domain.ShortURL{
		User:      user,
		OriginURL: pURL,
		ShortURL:  computeShortURL(pURL, user),
	})

	if err != nil {
		log.Print(err)
	}

	response(b.URLPrefix+shortURL, b, update)
}

func validateURL(pURL string) bool {
	parsedURL, err := url.Parse(pURL)
	if err != nil {
		log.Print(err)
		return false
	}

	// fix bug with cyrillic domains IDNA2003 https://datatracker.ietf.org/doc/html/rfc3490]
	hostname, err := idna.ToASCII(parsedURL.Host)
	_, err = net.DialTimeout("tcp", hostname+":http", time.Duration(2)*time.Second)
	if err != nil {
		_, err := net.DialTimeout("tcp", hostname+":https", time.Duration(2)*time.Second)
		if err != nil {
			log.Printf("Wrong address: %q", pURL)
			return false
		}
	}
	return true
}

func response(text string, b *Bot, update *tgbot.Update) {
	msg := tgbot.NewMessage(update.Message.Chat.ID, text)
	_, err := b.Bot.Send(msg)
	if err != nil {
		log.Print("Can't send message: ", err)
	}
}

func computeShortURL(originURL, user string) string {
	h := sha256.New()
	h.Write([]byte(originURL + user))
	hash := h.Sum(nil)
	b := base32.StdEncoding.EncodeToString(hash)
	return b[:5]
}
