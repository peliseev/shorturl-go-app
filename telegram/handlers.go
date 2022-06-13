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
	"strconv"
	"strings"
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

	CountMsgf = "По ссылке %q перешли %d раз%s."
)

func (b *Bot) HandleGreeting(update *tgbot.Update) {
	_, _ = response(fmt.Sprintf(GreetingFormat, update.Message.From.UserName), update.Message.Chat.ID, b)
}

func (b *Bot) HandleURL(update *tgbot.Update) {
	inMsg := update.Message.Text
	if !strings.HasPrefix(inMsg, "http") {
		inMsg = "https://" + inMsg
	}

	ok := validateURL(inMsg)
	if !ok {
		_, _ = response(ErrorMsg, update.Message.Chat.ID, b)
		return
	}
	user := update.Message.From.UserName
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	shortURL, err := b.URLService.CreateShortURL(ctx, &domain.ShortURL{
		User:      user,
		OriginURL: inMsg,
		ShortURL:  computeShortURL(inMsg, user),
	})
	if err != nil {
		log.Print(err)
	}
	responseWithCountButton(shortURL, b, update)
}

func (b *Bot) HandleCountButtonCallback(update *tgbot.Update) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	su, err := b.URLService.GetURLFollowCount(ctx, update.CallbackQuery.Data)
	if err != nil {
		log.Printf("error getting count for %q. err: %q", update.CallbackQuery.Data, err)
	}

	msg := fmt.Sprintf(CountMsgf, b.URLPrefix+su.ShortURL, su.Count, getA(su.Count))
	callbackCfg := tgbot.NewCallback(update.CallbackQuery.ID, "")

	_, _ = response(msg, update.CallbackQuery.Message.Chat.ID, b)
	_, err = b.Bot.AnswerCallbackQuery(callbackCfg)

	if err != nil {
		log.Print("error answering callback ", err)
	}

}

func validateURL(pURL string) bool {
	parsedURL, err := url.Parse(pURL)
	if err != nil {
		log.Print(err)
		return false
	}

	// fix bug with cyrillic domains IDNA2003 https://datatracker.ietf.org/doc/html/rfc3490]
	hostname, err := idna.ToASCII(parsedURL.Host)
	_, err = net.DialTimeout("tcp", hostname+":http", time.Duration(5)*time.Second)
	if err != nil {
		_, err := net.DialTimeout("tcp", hostname+":https", time.Duration(5)*time.Second)
		if err != nil {
			log.Printf("Invalid link: %q, domain %s is unreachable. Err: %q", pURL, hostname, err)
			return false
		}
	}
	return true
}

func responseWithCountButton(shortURL string, b *Bot, update *tgbot.Update) {
	sMsg, err := response(b.URLPrefix+shortURL, update.Message.Chat.ID, b)
	buttons := []tgbot.InlineKeyboardButton{
		tgbot.NewInlineKeyboardButtonData("Count", shortURL),
	}
	markUp := tgbot.NewInlineKeyboardMarkup(buttons)
	editMsg := tgbot.NewEditMessageReplyMarkup(update.Message.Chat.ID, sMsg.MessageID, markUp)
	_, err = b.Bot.Send(editMsg)
	if err != nil {
		log.Print("Can't send button: ", err)
	}
}

func response(text string, chatId int64, b *Bot) (tgbot.Message, error) {
	msg := tgbot.NewMessage(chatId, text)
	sMsg, err := b.Bot.Send(msg)
	if err != nil {
		log.Print("Can't send message: ", err)
		return tgbot.Message{}, err
	}
	return sMsg, err
}

func computeShortURL(originURL, user string) string {
	h := sha256.New()
	h.Write([]byte(originURL + user))
	hash := h.Sum(nil)
	b := base32.StdEncoding.EncodeToString(hash)
	return b[:5]
}

func getA(i int) string {
	s := strconv.Itoa(i)
	for _, a := range []string{"2", "3", "4"} {
		if (i < 5 || i > 21) && strings.HasSuffix(s, a) {
			return "а"
		}
	}
	return ""
}
