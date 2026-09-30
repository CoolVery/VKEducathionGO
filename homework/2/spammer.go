package main

import (
	"sync"
)

func RunPipeline(cmds ...cmd) {
	var wg sync.WaitGroup

	allCh := make([]chan interface{}, len(cmds)+1)

	for i := 0; i < len(cmds)+1; i++ {
		allCh[i] = make(chan interface{})
	}
	close(allCh[0])
	for index, command := range cmds {
		wg.Add(1)
		go func(in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			command(in, out)
		}(allCh[index], allCh[index+1])
	}
	
	wg.Wait()
}

func SelectUsers(in, out chan interface{}) {
	// 	in - string
	// 	out - User
	var email string

	for val := range in {
		email = val.(string)
	}

	res := GetUser(email)
	out <- res
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
}
