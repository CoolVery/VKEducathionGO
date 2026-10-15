package main

// сюда писать код

import (
	"context"
	"log"
	"os"

	"github.com/skinass/telegram-bot-api/v5"
)

var (
	// @BotFather в телеграме даст вам токен. Если захотите потыкать своего бота через телегу - используйте именно его
	BotToken = os.Getenv("BOTTOKEN")

	// Урл, в который будет стучаться телега при получении сообщения от пользователя.
	// Может быть как айпишником личной виртуалки, так и просто выдан сервисом для деплоя
	WebhookURL = os.Getenv("WEBHOOK")
)

func startTaskBot(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalln(CreateErrMsg(ErrLaunchBot, err))
	}
	return nil
	
}

func main() {
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
