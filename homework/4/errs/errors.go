package errs

import "errors"

var (
	ErrLaunchBot          = errors.New("не получилось запустить бота")
	ErrLaunchWebHook      = errors.New("не удалось создать вебхук")
	ErrSetWebHook         = errors.New("не удалось установить вебхук")
	ErrLaunchServer       = errors.New("не удалось запустить сервер")
	ErrLaunchEnv          = errors.New("не удалось обработать .env")
	ErrCmdNotFound        = errors.New("такой команды нет")
	ErrArgsIsNil          = errors.New("у команды не может быть пустные аргументы")
	ErrArdsIsNotNil       = errors.New("у команды не должно быть аргументов")
	ErrCtxDone            = errors.New("дейсвтие было отменено")
	ErrConvertStringToNum = errors.New("аргумент не получилось конвертировать в число")
	ErrHandleWrite        = errors.New("ендпоинт не смог отправить информацию")
	ErrNotBotSend         = errors.New("бот не смог отправить сообщение пользователю")
)

func CreateErrMsg(errClient, errInner error) string {
	return errClient.Error() + ": " + errInner.Error()
}
