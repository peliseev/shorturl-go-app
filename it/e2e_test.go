package it_test

import (
	"context"
	"fmt"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

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

type testSuits struct {
	testName  string
	originUrl string
	valid     bool
}

func testSuites() []testSuits {
	return []testSuits{
		{"valid url", "https://www.gojek.io/blog/golang-integration-testing-made-easy", true},
		{"russian site", "https://ицб.дом.рф/originators/securitization/for-what/", true},
		{"site without schema", "github.com", true},
		{"ip", "140.82.121.4", true},
		{"invalid url", "invalid url", false},
	}
}

type config struct {
	port, mongoUrl, urlPrefix string
}

type BotAPIMock struct {
	SentMsg string
}

func (b *BotAPIMock) Send(c tgbot.Chattable) (tgbot.Message, error) {
	switch c.(type) {
	case tgbot.MessageConfig:
		mc := c.(tgbot.MessageConfig)
		fmt.Println(mc.Text)
		b.SentMsg = mc.Text
	}
	return tgbot.Message{}, nil
}

func (b *BotAPIMock) GetUpdatesChan(tgbot.UpdateConfig) (tgbot.UpdatesChannel, error) {
	return nil, nil
}

func (b *BotAPIMock) AnswerCallbackQuery(tgbot.CallbackConfig) (tgbot.APIResponse, error) {
	return tgbot.APIResponse{}, nil
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

	// Tests
	t.Run("greeting", func(t *testing.T) { greetingTest(t, bot) })

	for _, ts := range testSuites() {
		t.Run(ts.testName, func(t *testing.T) { doTest(t, bot, ts.originUrl, ts.valid) })
	}

	t.Run("2 users with 1 link", func(t *testing.T) { twoUsersWithSameLink(t, bot, "https://go.dev/doc/effective_go") })
	t.Run("Test counter", func(t *testing.T) { testCounter(t, bot, "https://hh.ru/") })
}

func greetingTest(t *testing.T, bot telegram.Bot) {
	bot.HandleGreeting(prepareMessageUpdate("/start", TestUser))
	b := bot.Bot.(*BotAPIMock)
	if b.SentMsg != fmt.Sprintf(telegram.GreetingFormat, TestUser) {
		t.Errorf("Sent message: %q", b.SentMsg)
	}
}

func doTest(t *testing.T, bot telegram.Bot, originURL string, valid bool) {
	b := bot.Bot.(*BotAPIMock)
	bot.HandleURL(prepareMessageUpdate(originURL, TestUser))
	if valid {
		getAndCheckShortLink(t, b.SentMsg, originURL)
	} else {
		if b.SentMsg != telegram.ErrorMsg {
			t.Errorf("Sent message: %q", b.SentMsg)
		}
	}
}

func twoUsersWithSameLink(t *testing.T, bot telegram.Bot, originURL string) {
	b := bot.Bot.(*BotAPIMock)

	bot.HandleURL(prepareMessageUpdate(originURL, TestUser))
	msgToUser1 := b.SentMsg
	getAndCheckShortLink(t, msgToUser1, originURL)

	bot.HandleURL(prepareMessageUpdate(originURL, TestUser2))
	msgToUser2 := b.SentMsg
	getAndCheckShortLink(t, msgToUser1, originURL)

	if msgToUser1 == msgToUser2 {
		t.Errorf("User-1 short link: %q, User-2 short link: %q", msgToUser1, msgToUser2)
	}
}

func testCounter(t *testing.T, bot telegram.Bot, originURL string) {
	b := bot.Bot.(*BotAPIMock)

	bot.HandleURL(prepareMessageUpdate(originURL, TestUser))
	shortURL := b.SentMsg

	var expect string
	for i := 1; i <= 20; i++ {

		fmt.Printf("[%d] getAndCheckShortLink(*testing.T, %q, %q)\n", i, shortURL, originURL)
		getAndCheckShortLink(t, shortURL, originURL)
		bot.HandleCountButtonCallback(prepareButtonClickUpdate(shortURL[len(shortURL)-5:]))
		if i < 2 || i > 4 {
			expect = fmt.Sprintf(telegram.CountMsgf, shortURL, i, "")
		} else {
			expect = fmt.Sprintf(telegram.CountMsgf, shortURL, i, "а")
		}
		if b.SentMsg != expect {
			t.Errorf("Follow link %d times. Expect: %s, Got: %s\n", i, expect, b.SentMsg)
		}
	}
}

func prepareMessageUpdate(text, user string) *tgbot.Update {
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

func prepareButtonClickUpdate(shortURL string) *tgbot.Update {
	return &tgbot.Update{
		CallbackQuery: &tgbot.CallbackQuery{
			ID:   "5051",
			Data: shortURL,
			Message: &tgbot.Message{
				Chat: &tgbot.Chat{
					ID: 1337,
				},
			},
		},
	}
}

func getAndCheckShortLink(t *testing.T, shortURL, originURL string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(shortURL)
	if err != nil {
		t.Errorf("error while sending request to %s\n", shortURL)
	}
	if resp.StatusCode != 302 {
		t.Errorf("HTTP Status is '%d', Expected '%d'", resp.StatusCode, 302)
	}
	locationHeader := resp.Header.Get("Location")
	decodedValue, err := url.QueryUnescape(locationHeader)
	if !strings.HasSuffix(decodedValue, originURL) {
		t.Errorf("Header[Location] = %q and doesnt match %q", locationHeader, originURL)
	}
}

func envConfig(endpoint string) config {
	mongoUrl := "mongodb://" + endpoint
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(40000-1001+1) + 1001
	urlPrefix := "http://localhost:" + strconv.Itoa(port) + "/"

	return config{
		mongoUrl:  mongoUrl,
		urlPrefix: urlPrefix,
		port:      ":" + strconv.Itoa(port),
	}
}
