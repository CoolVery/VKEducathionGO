package commands

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type command struct {
	Name	string
	Sep byte
	Apply 	func(args string, userSend *tgbotapi.User) string
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