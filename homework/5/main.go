// Тут менять вам может потребоваться только коннект к базе
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	// DSN это соединение с базой
	// Вы можете изменить его на то, которое вам нужно
	// docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
	DSN = "root:1234@tcp(127.0.0.1:3306)/golang?parseTime=true&charset=latin1"
)

func main() {
	db, err := sql.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}
	var pingErr error
	for i := 1; i <= 15; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			log.Println("База данных готова к работе!")
			break
		}
	log.Printf("Ожидание базы данных (попытка %d/15): %v", i, pingErr)
	time.Sleep(2 * time.Second)
}

	handler, err := NewDBExplorer(db) //nolint:typecheck
	if err != nil {
		panic(err)
	}

	fmt.Println("starting server at :8082")
	if err := http.ListenAndServe(":8082", handler); err != nil {
		log.Printf("error listenAndServer: %v", err)
	}
}
