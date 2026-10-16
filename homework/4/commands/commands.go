package commands

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type command struct {
	Name	string
	WithArgs bool
	Apply 	func(args string, userSend *tgbotapi.User) string
}

var Commands = map[string]command {
	"/tasks": {
		Name: "/tasks",
		Apply: applyCmdTasks,
	},
	"/new": {
		Name: "/new",
		Apply: applyCmdNew,
	},
	"/assign": {
		Name: "/assign",
		WithArgs: true,
		Apply: applyCmdAssign,
	},
	"/unassign": {
		Name: "/unassign",
		WithArgs: true,
		Apply: applyCmdUnassign,
	},
	"/resolve": {
		Name: "/resolve",
		WithArgs: true,
		Apply: applyCmdResolve,
	},
	"/my": {
		Name: "/my",
		Apply: applyCmdMy,
	},
	"/owner": {
		Name: "/my",
		Apply: applyCmdOwner,
	},
}

func applyCmdTasks(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdNew(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdAssign(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdUnassign(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdResolve(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdMy(value string, userSend *tgbotapi.User) string {
	return "opa"
}
func applyCmdOwner(value string, userSend *tgbotapi.User) string {
	return "opa"
}