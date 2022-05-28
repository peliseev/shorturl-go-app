package telegram

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/url"
	"time"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/peliseev/shorturl-go-app/domain"
)

func (b *Bot) handleGreeting(update *tgbot.Update) {

	const greetingFormat = `Привет %s! 

	Этот бот умеет работать только с ссылками. 
	Прежде чем прислать ссылку, убедитесь в том, что:
	  - домен указан верно; 
	  - сайт онлайн;
	  - перед ссылкой нет лишних символов.`

	response(fmt.Sprintf(greetingFormat, update.Message.From.UserName), b, update)
}

func (b *Bot) handleURL(ctx context.Context, update *tgbot.Update) {
	const errorMsg = `
	Этот бот умеет работать только с ссылками. 
	Прежде чем прислать ссылку, убедитесь в том, что:
	  - домен указан верно; 
	  - сайт онлайн;
	  - перед ссылкой нет лишних символов.`

	pURL := update.Message.Text
	ok := validateURL(pURL)

	if !ok {
		response(errorMsg, b, update)
		return
	}

	user := update.Message.From.UserName
	shortURL, err := b.urlService.CreateShortURL(ctx, &domain.ShortURL{
		User:      user,
		OriginURL: pURL,
		ShortURL:  computeShortURL(pURL, user),
	})

	if err != nil {
		log.Print(err)
	}

	response(shortURL, b, update)
}

func validateURL(pURL string) bool {
	url, err := url.Parse(pURL)
	if err != nil {
		log.Print(err)
		return false
	}
	_, err = net.DialTimeout("tcp", url.Host+":http", time.Duration(5)*time.Second)
	if err != nil {
		_, err := net.DialTimeout("tcp", url.Host+":https", time.Duration(5)*time.Second)
		if err != nil {
			log.Printf("Site %s is unreachable", pURL)
			return false
		}
	}
	return true
}

func response(text string, b *Bot, update *tgbot.Update) {
	msg := tgbot.NewMessage(update.Message.Chat.ID, text)
	_, err := b.bot.Send(msg)
	if err != nil {
		log.Print("Can't send message: ", err)
	}
}

func computeShortURL(originURL, user string) string {
	h := sha256.New()
	h.Write([]byte(originURL + user))
	hash := h.Sum(nil)
	b := base64.StdEncoding.EncodeToString(hash)
	return b[:5]
}
