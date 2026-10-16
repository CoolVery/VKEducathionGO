package commands

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type command struct {
	Name	string
	Apply 	func(value string, userSend *tgbotapi.User) string
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
		Apply: applyCmdAssign,
	},
	"/unassign": {
		Name: "/unassign",
		Apply: applyCmdUnassign,
	},
	"/resolve": {
		Name: "/resolve",
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