package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/skinass/telegram-bot-api/v5"

	"hm4/commands"
	. "hm4/errors"
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
		if _, err := w.Write([]byte("Server alive")); err != nil {
			log.Fatalln(CreateErrMsg(ErrHandleWrite, err))
		}
	})

	go func() {
		log.Fatalln(CreateErrMsg(ErrLaunchServer, http.ListenAndServe(":3000", nil)))
	}()
	log.Println("Start server")
	
	for message := range newMessages {
		log.Println(message.Message.Chat.ID, " | ", message.Message.From.UserName, " | ", message.Message.Text)
		text := message.Message.Text

		// 1. Грубо вытащить имя команды — только чтобы найти её в мапе.
		head := ""
		if idx := strings.IndexAny(text, "_ "); idx != -1 {
		 	head = text[:idx]
		}

		// 2. Найти команду.
		cmd, ok := commands.Commands[head]
		if !ok {
		    if _, err := bot.Send(tgbotapi.NewMessage(message.Message.Chat.ID, ErrCmdNotFound.Error())); err != nil {
				log.Fatalln(CreateErrMsg(ErrNotBotSend, err))
			}
		    continue
		}

		// 3. Строго проверить формат и достать аргумент.
		var args string
		if cmd.Sep == 0 {
		    // Аргументов быть не должно — текст обязан точно совпасть с именем.
		    if text != cmd.Name {
				if _, err := bot.Send(tgbotapi.NewMessage(message.Message.Chat.ID, ErrCmdNotFound.Error())); err != nil {
					log.Fatalln(CreateErrMsg(ErrNotBotSend, err))
				}
	        	continue
		 	}
		} else {
		    // Ожидаемый префикс: имя + разделитель.
		    prefix := cmd.Name + string(cmd.Sep)
		    if !strings.HasPrefix(text, prefix) {
		        if _, err := bot.Send(tgbotapi.NewMessage(message.Message.Chat.ID, ErrCmdNotFound.Error())); err != nil {
					log.Fatalln(CreateErrMsg(ErrNotBotSend, err))
				}
	        	continue
		    }
		    // Всё после префикса — аргумент.
		    args = strings.TrimSpace(text[len(prefix):])
		}

		// 4. Вызвать обработчик и отправить ответ.
		answer := cmd.Apply(ctx, args, message.Message.From)
		for _, out := range answer {
			if _, err := bot.Send(tgbotapi.NewMessage(out.ChatID, out.Text)); err != nil {
				log.Fatalln(CreateErrMsg(ErrNotBotSend, err))
			}
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
