package main

// сюда писать код

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/skinass/telegram-bot-api/v5"
)

// var (
// 	// @BotFather в телеграме даст вам токен. Если захотите потыкать своего бота через телегу - используйте именно его
// 	BotToken = os.Getenv("BOTTOKEN")

// 	// Урл, в который будет стучаться телега при получении сообщения от пользователя.
// 	// Может быть как айпишником личной виртуалки, так и просто выдан сервисом для деплоя
// 	// для serveo туннируем - ssh -R 80:localhost:3000 serveo.net
// 	WebhookURL = os.Getenv("WEBHOOK")
// )

func startTaskBot(ctx context.Context) error {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(CreateErrMsg(ErrLaunchEnv, err))
	}
	botToken := os.Getenv("BOTTOKEN")
	webHook := os.Getenv("WEBHOOK")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalln(CreateErrMsg(ErrLaunchBot, err))
	}

	wh, err := tgbotapi.NewWebhook(webHook)
	if err != nil {
		log.Fatalln(CreateErrMsg(ErrLaunchWebHook, err))
	}
	
	_, err = bot.Request(wh)
	if err != nil {
		log.Fatalln(CreateErrMsg(ErrSetWebHook, err))
	}
	
	newMessages := bot.ListenForWebhook("/")

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server alive"))
	})

	go func() {
		log.Fatalln(CreateErrMsg(ErrLaunchServer, http.ListenAndServe(":3000", nil)))
	}()
	log.Println("Setart server")
	
	for message := range newMessages {
		msg := tgbotapi.NewMessage(
			message.Message.Chat.ID,
			"Hey",
		)
		bot.Send(msg)
	}
	return nil

}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
