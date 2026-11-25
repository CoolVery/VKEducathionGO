package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/skinass/telegram-bot-api/v5"
	"hm4/commands"
	"hm4/errs"
)

func loadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("%s", errs.CreateErrMsg(errs.ErrLaunchEnv, err))
	}
	return nil
}

func loadBot(botToken string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("%s", errs.CreateErrMsg(errs.ErrLaunchBot, err))
	}
	return bot, nil
}

func loadWebHook(webHook string) (tgbotapi.WebhookConfig, error) {
	wh, err := tgbotapi.NewWebhook(webHook)
	if err != nil {
		return tgbotapi.WebhookConfig{}, fmt.Errorf("%s", errs.CreateErrMsg(errs.ErrLaunchWebHook, err))
	}
	return wh, err
}

func loadSrv() {
	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Server alive")); err != nil {
			log.Fatalln(errs.CreateErrMsg(errs.ErrHandleWrite, err))
		}
	})
	srv := &http.Server{
		Addr:         ":3000",
		Handler:      nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() {
		log.Fatalln(errs.CreateErrMsg(errs.ErrLaunchServer, srv.ListenAndServe()))
	}()
	log.Println("Start server")
}

func sendToUser(bot *tgbotapi.BotAPI, chatId int64, text string) error {
	if _, err := bot.Send(tgbotapi.NewMessage(chatId, text)); err != nil {
		return fmt.Errorf("%s", errs.CreateErrMsg(errs.ErrNotBotSend, err))
	}
	return nil
}

func handleMessage(ctx context.Context, bot *tgbotapi.BotAPI, msg *tgbotapi.Message) error {
	log.Println(msg.Chat.ID, " | ", msg.From.UserName, " | ", msg.Text)

	cmd, args, err := commands.ParseCommand(msg.Text)
	if err != nil {
		if err := sendToUser(bot, msg.Chat.ID, err.Error()); err != nil {
			return err
		}
	}

	for _, out := range cmd.Apply(ctx, args, msg.From) {
		if err := sendToUser(bot, out.ChatID, out.Text); err != nil {
			return err
		}
	}
	return nil
}

func startTaskBot(ctx context.Context) error {
	if err := loadEnv(); err != nil {
		return err
	}

	botToken := os.Getenv("BOTTOKEN")
	webHook := os.Getenv("WEBHOOK")

	bot, err := loadBot(botToken)
	if err != nil {
		return err
	}
	wh, err := loadWebHook(webHook)
	if err != nil {
		return err
	}

	_, err = bot.Request(wh)
	if err != nil {
		return fmt.Errorf("%s", errs.CreateErrMsg(errs.ErrSetWebHook, err))
	}

	newMessages := bot.ListenForWebhook("/")

	loadSrv()

	for message := range newMessages {
		if err := handleMessage(ctx, bot, message.Message); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
