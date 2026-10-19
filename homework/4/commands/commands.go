package commands

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"hm4/errs"
	"hm4/modules"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

type outgoing struct {
	ChatID int64
	Text   string
}

type command struct {
	Name  string
	Sep   byte
	Apply func(ctx context.Context, args string, userSend *tgbotapi.User) []outgoing
}

var Commands = map[string]command{
	"/tasks": {
		Name:  "/tasks",
		Apply: applyCmdTasks,
	},
	"/new": {
		Name:  "/new",
		Sep:   ' ',
		Apply: applyCmdNew,
	},
	"/assign": {
		Name:  "/assign",
		Sep:   '_',
		Apply: applyCmdAssign,
	},
	"/unassign": {
		Name:  "/unassign",
		Sep:   '_',
		Apply: applyCmdUnassign,
	},
	"/resolve": {
		Name:  "/resolve",
		Sep:   '_',
		Apply: applyCmdResolve,
	},
	"/my": {
		Name:  "/my",
		Apply: applyCmdMy,
	},
	"/owner": {
		Name:  "/owner",
		Apply: applyCmdOwner,
	},
}

func ParseCommand(text string) (command, string, error) {
	head := text
	if idx := strings.IndexAny(text, "_ "); idx != -1 {
		head = text[:idx]
	}

	cmd, ok := Commands[head]
	if !ok {
		return command{}, "", errs.ErrCmdNotFound
	}

	if cmd.Sep == 0 {
		if text != cmd.Name {
			return command{}, "", errs.ErrCmdNotFound
		}
		return cmd, "", nil
	}

	prefix := cmd.Name + string(cmd.Sep)
	if !strings.HasPrefix(text, prefix) {
		return command{}, "", errs.ErrCmdNotFound
	}
	return cmd, strings.TrimSpace(text[len(prefix):]), nil
}

func applyCmdTasks(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value != "" {
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrCtxDone.Error()}}
	default:
	}

	if len(modules.AllTasks) == 0 {
		return []outgoing{{userSend.ID, errs.ErrNotTasks.Error()}}
	}
	
	var printAllTask []string
	for _, task := range modules.AllTasks {
		switch task.StatusCode {
		case modules.NotAssigned:
			stringTask := fmt.Sprintf(`%d. %s by @%s
/assign_%d`, task.ID, task.Description, task.Author.UserName, task.ID)
			printAllTask = append(printAllTask, stringTask)
		case modules.Assigned:
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
		return []outgoing{{userSend.ID, errs.ErrArgsIsNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrCtxDone.Error()}}
	default:
	}

	newTask := modules.NewTask(modules.MaxId, value, userSend)
	modules.MaxId++
	modules.AllTasks = append(modules.AllTasks, newTask)
	result := fmt.Sprintf("Задача \"%s\" создана, id=%d", value, newTask.ID)
	return []outgoing{{userSend.ID, result}}
}

func applyCmdAssign(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value == "" {
		return []outgoing{{userSend.ID, errs.ErrArgsIsNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	default:
	}
	valueNum, err := strconv.Atoi(value)
	if err != nil {
		return []outgoing{{userSend.ID, errs.ErrConvertStringToNum.Error()}}
	}
	var allOutgoing []outgoing
	for i := range modules.AllTasks {
		if modules.AllTasks[i].ID == valueNum {
			task := &modules.AllTasks[i]
			switch task.StatusCode {
			case modules.NotAssigned:
				if task.Author.ID != userSend.ID {
					msgToAuthor := fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Description, userSend.UserName)
					allOutgoing = append(allOutgoing, outgoing{task.Author.ID, msgToAuthor})
				}
				msgToExecutor := fmt.Sprintf(`Задача "%s" назначена на вас`, task.Description)
				task.Executor = userSend
				task.StatusCode = modules.Assigned
				allOutgoing = append(allOutgoing, outgoing{userSend.ID, msgToExecutor})
				return allOutgoing
			case modules.Assigned:
				resultToExecutor := fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Description, userSend.UserName)
				allOutgoing = append(allOutgoing, outgoing{task.Executor.ID, resultToExecutor})
				resultToNewExecutor := fmt.Sprintf(`Задача "%s" назначена на вас`, task.Description)
				allOutgoing = append(allOutgoing, outgoing{userSend.ID, resultToNewExecutor})
				task.Executor = userSend
				return allOutgoing

			}
		}
	}
	return []outgoing{{userSend.ID, errs.ErrTaskNotFound.Error()}}
}

func applyCmdUnassign(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value == "" {
		return []outgoing{{userSend.ID, errs.ErrArgsIsNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	default:
	}

	valueNum, err := strconv.Atoi(value)
	if err != nil {
		return []outgoing{{userSend.ID, errs.ErrConvertStringToNum.Error()}}
	}

	for i := range modules.AllTasks {
		if modules.AllTasks[i].ID == valueNum {
			task := &modules.AllTasks[i]
			if task.StatusCode == modules.NotAssigned {
				return []outgoing{{userSend.ID, errs.ErrTryUnassignedNotAssignedTask.Error()}}
			}
			if task.Executor.ID != userSend.ID {
				return []outgoing{{userSend.ID, errs.ErrTryUnassignedNotYourTask.Error()}}
			}

			resultOut := make([]outgoing, 0, 2)
			msgToSender := "Принято"
			resultOut = append(resultOut, outgoing{userSend.ID, msgToSender})
			task.StatusCode = modules.NotAssigned
			task.Executor = nil
			msgToAuthor := fmt.Sprintf(`Задача "%s" осталась без исполнителя`, task.Description)
			resultOut = append(resultOut, outgoing{task.Author.ID, msgToAuthor})
			return resultOut
		}
	}
	return []outgoing{{userSend.ID, errs.ErrTaskNotFound.Error()}}
}

func applyCmdResolve(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value == "" {
		return []outgoing{{userSend.ID, errs.ErrArgsIsNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	default:
	}

	valueNum, err := strconv.Atoi(value)
	if err != nil {
		return []outgoing{{userSend.ID, errs.ErrConvertStringToNum.Error()}}
	}
	for i := range modules.AllTasks {
		if modules.AllTasks[i].ID == valueNum {
			task := modules.AllTasks[i]

			if task.StatusCode == modules.NotAssigned {
				return []outgoing{{userSend.ID,errs.ErrTryResolveNotAssignedTask.Error()}}
			}
			if userSend.ID != task.Executor.ID {
				return []outgoing{{userSend.ID,errs.ErrTryResolveNotYourTask.Error()}}
			}
			resultOut := make([]outgoing, 0, 2)
			modules.AllTasks[i] = modules.AllTasks[len(modules.AllTasks) - 1]
			modules.AllTasks = modules.AllTasks[:len(modules.AllTasks) - 1]
			msgSendUser := fmt.Sprintf(`Задача "%s" выполнена`, task.Description) 
			msgSendAuthor := fmt.Sprintf(`Задача "%s" выполнена @%s`, task.Description, task.Executor.UserName) 
			resultOut = append(resultOut, outgoing{userSend.ID, msgSendUser}, outgoing{task.Author.ID, msgSendAuthor})	
			return resultOut	
		}	
	}
	return []outgoing{{userSend.ID, errs.ErrTaskNotFound.Error()}}
}

func applyCmdMy(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value != "" {
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrCtxDone.Error()}}
	default:
	}

	if len(modules.AllTasks) == 0 {
		return []outgoing{{userSend.ID, errs.ErrNotTasks.Error()}}
	}

	var outStringWithTask []string
	for i := range modules.AllTasks {
		task := modules.AllTasks[i]
		if modules.AllTasks[i].Executor.ID == userSend.ID {
			msg := fmt.Sprintf(`%d. %s by @%s
/unassign_%d /resolve_%d`, task.ID, task.Description, task.Author.UserName, task.ID, task.ID)
			outStringWithTask = append(outStringWithTask, msg)
		}
	}
	if len(outStringWithTask) == 0 {
		return []outgoing{{userSend.ID, errs.ErrTaskNotFound.Error()}}
	}
	
	resultMsg := strings.Join(outStringWithTask,  "\n\n")
	return []outgoing{outgoing{userSend.ID, resultMsg}}
}

func applyCmdOwner(ctx context.Context, value string, userSend *tgbotapi.User) []outgoing {
	if value != "" {
		return []outgoing{{userSend.ID, errs.ErrArdsIsNotNil.Error()}}
	}

	select {
	case <-ctx.Done():
		return []outgoing{{userSend.ID, errs.ErrCtxDone.Error()}}
	default:
	}

	if len(modules.AllTasks) == 0 {
		return []outgoing{{userSend.ID, errs.ErrNotTasks.Error()}}
	}
	var outStringWithTask []string
	for i := range modules.AllTasks {
		task := modules.AllTasks[i]
		if modules.AllTasks[i].Author.ID == userSend.ID {
			msg := fmt.Sprintf(`%d. %s by @%s
/assign_%d`, task.ID, task.Description, task.Author.UserName, task.ID)
			outStringWithTask = append(outStringWithTask, msg)
		}
	}
	if len(outStringWithTask) == 0 {
		return []outgoing{{userSend.ID, errs.ErrTaskNotFound.Error()}}
	}
	
	resultMsg := strings.Join(outStringWithTask,  "\n\n")
	return []outgoing{outgoing{userSend.ID, resultMsg}}
}
