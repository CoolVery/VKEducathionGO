package main

import "errors"

var (
	ErrLaunchBot = errors.New("Не получилось запустить бота")
	ErrLaunchWebHook = errors.New("Не удалось создать вебхук")
	ErrSerWebHook = errors.New("Не удалось установить вебхук")
	
)

func CreateErrMsg(errClient, errInner error) string {
	return errClient.Error() + ": " + errInner.Error()
}