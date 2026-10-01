package main

import (
	"sync"
)
//Глобальный буфер для антибрута, сделано так, потому что менять сигнатуру фуннкций нельзя
//Из себя этот буфер представляет буфферизированный канал, он будет копить в себе запросы
//и блокироваться, если переполнен
var buf = make(chan struct{}, HasSpamMaxAsyncRequests)

func RunPipeline(cmds ...cmd) {
	var wg sync.WaitGroup
	//Сеть для cmd, которые будут связывать все функции
	allCh := make([]chan interface{}, len(cmds)+1)
	//Инициализируем все каналы
	for i := 0; i < len(cmds)+1; i++ {
		allCh[i] = make(chan interface{})
	}
	//Закрываем первый канал, потому что первая команда читает не из канала, а из других источников
	close(allCh[0])
	for index, command := range cmds {
		//Добавляем в группу
		wg.Add(1)
		//Запускаем команду, передавая каналы из Сети
		go func(in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			command(in, out)
		}(allCh[index], allCh[index+1])
	}
	for range allCh[len(cmds)]{}

	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	var wg sync.WaitGroup
	var mu sync.Mutex
	//Мапа кэша для отправленных юзеров
	cashMap := make(map[User]struct{})
	//Обходим все значения в канале in
	for email := range in {
		wg.Add(1)
		go func(e string) {
			defer wg.Done()
			//Делаем запрос и получаем нашего юзера по почте
			us := GetUser(e)
			mu.Lock()
			//Если в кэше нет нашего юзер есть, то мы делаем запись и после отправялем его в out
			//Если в кэше есть юзер, то мы ничего не делаем, ведь его уже отправили
			_, ok := cashMap[us]
			if !ok {
				cashMap[us] = struct{}{}
			}
			mu.Unlock()
			if !ok {
				out <- us
			} 
		}(email.(string))
	}
	wg.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	var wg sync.WaitGroup
	//Создаем слайс для хранения юзеров
	users := make([]User, 0, GetMessagesMaxUsersBatch)
	//Проходим по кадому Юзеру
	for val := range in {
		//Получаем юзера
		user := val.(User)
		//Если в батче не макс.количество, то добавляем
		if len(users) != GetMessagesMaxUsersBatch {
			users = append(users, user)
		} else {
			wg.Add(1)
			//Иначе копируем батч, чтобы копию передать в горутину, а с главный буфер сбрасываем до 0
			tempUsers := make([]User, len(users))
			copy(tempUsers, users)
			//Сбросили
			users = users[:0]
			//Добавили текущего пользователя
			users = append(users, user)
			go func(u []User) {
				defer wg.Done()
				mess, err := GetMessages(u...)
				//Отправляем в канал все сообщения из слайса
				if err == nil {
					for _, message := range mess {
						out <- message 
					}
				}
			//Передаем копию
			}(tempUsers)
		}
	}
	//Проверка, что если в канале больше нет данных, но в буфере что-то осталось, то их тоже надо отработать
	//Один раз запускаем горутину с остатками данных
	if len(users) != 0 {
		wg.Add(1)
		go func(u []User) {
				defer wg.Done()
				mess, err := GetMessages(u...)
				if err == nil {
					for _, message := range mess {
						out <- message 
					}
				}
		}(users)
	}
	wg.Wait()
}

 

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	var wg sync.WaitGroup
	//Проходим по каждому сообщению
	for val := range in {
		//Получаем сообщение
		msg := val.(MsgID)
		wg.Add(1)
		go func(m MsgID) {
			//Сначала регистрируем освобождение слота в буффере от антибрута, чтобы он точно выполнился
			//После выполнения работы слот из буфера прочтется и освободится
			defer func() { <-buf }()
			//После wg.Done
			defer wg.Done()
			//Занимаем слот в буфере
			buf <- struct{}{}
			//Делаем запрос
			res, err := HasSpam(m)
			if err == nil {
				//Создаем объект и отправляем
				msgData := MsgData{
					ID: m,
					HasSpam: res,
				}
				out <- msgData
			}
		}(msg)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
