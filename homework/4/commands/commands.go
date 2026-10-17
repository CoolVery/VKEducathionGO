package commands

import (
	"context"
	"fmt"
	. "hm4/errors"
	. "hm4/modules"
	"strconv"
	"strings"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type outgoing struct {
    ChatID int64
    Text   string
}

type command struct {
	Name	string
	Sep byte
	Apply 	func(ctx context.Context, args string, userSend *tgbotapi.User) []outgoing
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

func applyCmdTasks(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value != "" {
		return []outgoing{{userSend.ID, ErrArdsIsNotNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, ErrCtxDone.Error()}}
	default:
	}

	var printAllTask []string
	for _, task := range AllTasks {
		switch task.StatusCode {
		case NotAssigned:
			stringTask := fmt.Sprintf(`%d. %s by @%s
/assign_%d`, task.ID, task.Description, task.Author.UserName, task.ID)
			printAllTask = append(printAllTask, stringTask)
		case Assigned:
			var stringTask string
			if task.Executor.ID == userSend.ID {
				stringTask = fmt.Sprintf(`%d. %s by @%s
assignee: я
/unassign_%d /resolve_%d`, task.ID, task.Description, task.Author.UserName, task.ID, task.ID)
			} else {
			stringTask = fmt.Sprintf(`%d. %s by @%s
assignee: @%s`, task.ID, task.Description, task.Author.UserName, task.Executor.UserName)
			}
			printAllTask = append(printAllTask, stringTask)
		}
	}
	result := strings.Join(printAllTask, "\n\n")
	return []outgoing{{userSend.ID, result}}
}
func applyCmdNew(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value == "" {
		return []outgoing{{userSend.ID, ErrArgsIsNil.Error()}}
	}

	select {
    case <-ctx.Done():
		return []outgoing{{userSend.ID, ErrCtxDone.Error()}}
    default:
    }

	newTask := NewTask(MaxId, value, userSend)
	MaxId++
	AllTasks = append(AllTasks, newTask)
	result := fmt.Sprintf("Задача \"%s\" создана, id=%d", value, newTask.ID)
	return []outgoing{{userSend.ID, result}}
}

func applyCmdAssign(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value == "" {
		return []outgoing{{userSend.ID, ErrArgsIsNil.Error()}}
	}

	select {
	case <- ctx.Done():
		return []outgoing{{userSend.ID, ErrArdsIsNotNil.Error()}}
	default:
	}
	valueNum, err := strconv.Atoi(value)
	if err != nil {
		return []outgoing{{userSend.ID, ErrConvertStringToNum.Error()}}
	}
	var result string
	var allOutgoing []outgoing
	for _, task := range AllTasks {
		if task.ID == valueNum {
			switch task.Executor == nil {
			case true:
				if task.Author.ID != userSend.ID {
					result = fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Description, userSend.UserName)
					allOutgoing = append(allOutgoing, outgoing{task.Author.ID, result})	
				} 
				result = fmt.Sprintf(`Задача "%s" назначена на вас`, task.Description)
				allOutgoing = append(allOutgoing, outgoing{userSend.ID, result})
			}
		}
	}
		return allOutgoing
}
func applyCmdUnassign(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
		return []outgoing{{userSend.ID, "result"}}
}
func applyCmdResolve(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
		return []outgoing{{userSend.ID, "result"}}
}
func applyCmdMy(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
		return []outgoing{{userSend.ID, "result"}}
}
func applyCmdOwner(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
		return []outgoing{{userSend.ID, "result"}}
}