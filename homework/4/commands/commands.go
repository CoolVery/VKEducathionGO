package commands

import (
	"context"
	"fmt"
	. "hm4/errors"
	. "hm4/modules"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type command struct {
	Name	string
	Sep byte
	Apply 	func(ctx context.Context, args string, userSend *tgbotapi.User) string
}

var Commands = map[string]command {
	"/tasks": {
		Name: "/tasks",
		Apply: applyCmdTasks,
	},
	"/new": {
		Name: "/new",
		Sep: ' ',
		Apply: applyCmdNew,
	},
	"/assign": {
		Name: "/assign",
		Sep: '_',
		Apply: applyCmdAssign,
	},
	"/unassign": {
		Name: "/unassign",
		Sep: '_',
		Apply: applyCmdUnassign,
	},
	"/resolve": {
		Name: "/resolve",
		Sep: '_',
		Apply: applyCmdResolve,
	},
	"/my": {
		Name: "/my",
		Apply: applyCmdMy,
	},
	"/owner": {
		Name: "/owner",
		Apply: applyCmdOwner,
	},
}

func applyCmdTasks(ctx context.Context, value string, userSend *tgbotapi.User) string {
	if value != "" {
		return ErrArdsIsNotNil.Error()
	}
	return "opa"
}
func applyCmdNew(ctx context.Context, value string, userSend *tgbotapi.User) string {
	if value == "" {
		return ErrArgsIsNil.Error()
	}

	select {
    case <-ctx.Done():
        return ErrCtxDone.Error()
    default:
    }

	idNewTask := len(AllTasks) + 1
	newTask := NewTask(idNewTask, value, userSend)
	AllTasks = append(AllTasks, newTask)
	return fmt.Sprintf("Задача \"%s\" создана, id=%d", value, idNewTask)
}

func applyCmdAssign(ctx context.Context, value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdUnassign(ctx context.Context, value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdResolve(ctx context.Context, value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdMy(ctx context.Context, value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdOwner(ctx context.Context, value string, userSend *tgbotapi.User) string {
	return "opa"
}