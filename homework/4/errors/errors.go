package errors

import "errors"

var (
	ErrLaunchBot = errors.New("Не получилось запустить бота")
	ErrLaunchWebHook = errors.New("Не удалось создать вебхук")
	ErrSetWebHook = errors.New("Не удалось установить вебхук")
	ErrLaunchServer = errors.New("Не удалось запустить сервер")
	ErrLaunchEnv = errors.New("Не удалось обработать .env")
	ErrCmdNotFound = errors.New("Такой команды нет")
	ErrArgsIsNil = errors.New("У команды не может быть пустные аргументы")
	ErrArdsIsNotNil = errors.New("У команды не должно быть аргументов")
	ErrCtxDone = errors.New("Дейсвтие было отменено")
)

func CreateErrMsg(errClient, errInner error) string {
	return errClient.Error() + ": " + errInner.Error()
}