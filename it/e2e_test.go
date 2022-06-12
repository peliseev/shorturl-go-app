package it_test

import (
	"context"
	"fmt"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"net/http"
	"net/url"
	"testing"

	"github.com/peliseev/shorturl-go-app/mongo"
	"github.com/peliseev/shorturl-go-app/server"
	"github.com/peliseev/shorturl-go-app/telegram"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	TestUser  = "test_user"
	TestUser2 = "test_user_2"
)

type config struct {
	port, mongoUrl, urlPrefix string
}

type BotAPIMock struct {
	SentMsg string
}

func (b *BotAPIMock) Send(c tgbot.Chattable) (tgbot.Message, error) {
	mc := c.(tgbot.MessageConfig)
	fmt.Println(mc.Text)
	b.SentMsg = mc.Text
	return tgbot.Message{}, nil
}

func (b *BotAPIMock) GetUpdatesChan(tgbot.UpdateConfig) (tgbot.UpdatesChannel, error) {
	return nil, nil
}

func TestWithMongoDB(t *testing.T) {
	ctx := context.Background()
	req := tc.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	mongoC, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Error("error starting mongoDB container", err)
	}
	defer func(mongoC tc.Container, ctx context.Context) {
		err := mongoC.Terminate(ctx)
		if err != nil {
			panic("WARNING!!! mongo container steel alive!")
		}
	}(mongoC, ctx)

	endpoint, err := mongoC.Endpoint(ctx, "")
	if err != nil {
		t.Fatal("error getting mongoDB container endpoint", err)
	}

	cfg := envConfig(endpoint)
	db := mongo.Open(cfg.mongoUrl)
	service := mongo.NewShortURLService(db)
	testServer := server.NewServer(db, service)
	bot := telegram.Bot{
		URLPrefix:  cfg.urlPrefix,
		URLService: service,
		Bot:        &BotAPIMock{},
	}

	go func() {
		err := testServer.Run(cfg.port)
		if err != nil {
			t.Error("error starting testServer", err)
			return
		}
	}()

	t.Run("greeting", func(t *testing.T) { greetingTest(t, bot) })
	t.Run("valid url", func(t *testing.T) { validUrl(t, bot) })
	t.Run("invalid url", func(t *testing.T) { invalidUrl(t, bot) })
	t.Run("2 users with 1 link", func(t *testing.T) { twoUsersWithSameLink(t, bot) })
	t.Run("russian site", func(t *testing.T) { russianSite(t, bot) })
}

func greetingTest(t *testing.T, bot telegram.Bot) {
	bot.HandleGreeting(prepareMsg("/start", TestUser))
	b := bot.Bot.(*BotAPIMock)
	if b.SentMsg != fmt.Sprintf(telegram.GreetingFormat, TestUser) {
		t.Errorf("Sent message: %q", b.SentMsg)
	}
}

func validUrl(t *testing.T, bot telegram.Bot) {
	originURL := "https://www.gojek.io/blog/golang-integration-testing-made-easy"
	bot.HandleURL(prepareMsg(originURL, TestUser))

	b := bot.Bot.(*BotAPIMock)
	if b.SentMsg != "http://localhost:8080/IEGRD" {
		t.Errorf("\nSentMsg = %s\nExpected = http://localhost:8080/IEGRD\n", b.SentMsg)
	}

	getShortLink(t, b.SentMsg, originURL)
}

func invalidUrl(t *testing.T, bot telegram.Bot) {
	bot.HandleURL(prepareMsg("https://www.pasddsadsadsadsa.io/blog/golang-integration-testing-made-easy",
		TestUser))

	b := bot.Bot.(*BotAPIMock)
	if b.SentMsg != telegram.ErrorMsg {
		t.Errorf("Sent message: %q", b.SentMsg)
	}
}

func twoUsersWithSameLink(t *testing.T, bot telegram.Bot) {
	b := bot.Bot.(*BotAPIMock)

	bot.HandleURL(prepareMsg("https://diablo4.blizzard.com/en-us/", TestUser))
	msgToUser1 := b.SentMsg

	bot.HandleURL(prepareMsg("https://diablo4.blizzard.com/en-us/", TestUser2))
	msgToUser2 := b.SentMsg

	if msgToUser1 == msgToUser2 {
		t.Errorf("User-1 short link: %q, User-2 short link: %q", msgToUser1, msgToUser2)
	}
}

func russianSite(t *testing.T, bot telegram.Bot) {
	originURL := "https://ицб.дом.рф/originators/securitization/for-what/"
	b := bot.Bot.(*BotAPIMock)
	bot.HandleURL(prepareMsg(originURL, TestUser))

	getShortLink(t, b.SentMsg, originURL)
}

func prepareMsg(text, user string) *tgbot.Update {
	return &tgbot.Update{
		Message: &tgbot.Message{
			Text: text,
			From: &tgbot.User{
				UserName: user,
			},
			Chat: &tgbot.Chat{
				ID: 1337,
			},
		},
	}
}

func getShortLink(t *testing.T, shortURL, originURL string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(shortURL)
	if err != nil {
		t.Errorf("error while sending request to %s", shortURL)
	}
	if resp.StatusCode != 302 {
		t.Errorf("HTTP Status is '%d', Expected '%d'", resp.StatusCode, 302)
	}
	locationHeader := resp.Header.Get("Location")
	decodedValue, err := url.QueryUnescape(locationHeader)
	if decodedValue != originURL {
		t.Errorf("Header[Location] = %q and doesnt match %q", locationHeader, originURL)
	}
}

func envConfig(endpoint string) config {
	mongoUrl := "mongodb://" + endpoint
	urlPrefix := "http://localhost:8080/"
	port := ":8080"

	return config{
		mongoUrl:  mongoUrl,
		urlPrefix: urlPrefix,
		port:      port,
	}
}
