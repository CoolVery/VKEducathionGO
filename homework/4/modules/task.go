package modules

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type Task struct {
	ID 				int
	Author     	*tgbotapi.User
	Description 	string
	Executor		*tgbotapi.User
	StatusCode		int
}

func NewTask(idNewTask int, description string, author *tgbotapi.User) Task {
	newTask := Task{
		ID: idNewTask,
		Author: author,
		Description: description,
		StatusCode: NotAssigned,
	}
	return newTask
}
