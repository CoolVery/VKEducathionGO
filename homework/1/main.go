package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//Игрок
type Player struct {
	//Инвентарь игрока
	Inventory map[string]*Item
	//Инвентарь доступен
	IsInventoryAccess bool
	//Текущая локация пользователя
	CurrentLocathion *Room
}
//Контейнер (на столе, на стуле)
type Container struct {
	//Имя контейнера
	Name string
	//Предметы, которые хранятся в контейнере
	InternalItems []*Item
}
//Таргеты в комнатах, которые сопаствялются с определенными предметами и выдыют результат
// [Дверь] - ключи, [Ключи] - дверь открыта 
type Target struct {
	//Имя таргета
	Name string
	//Мапа сопастовления имени таргета с предметом
	ItemsApply map[string]*Item
	//Мапа сопастовления предмета и результата использования с таргетом
	ApplyResult map[*Item]string
}
//Комнаты
type Room struct {
	//Имя
	Name string
	//Описание, которое используется для формирования строки команды "Идти"
	Description string
	//Уникальный текст в начале, которое используется для формирования строки команды "Осмотреться" 
	UniqueTextInStart string
	//Уникальный текст в конце, которое используется для формирования строки команды "Осмотреться" 
	UniqueTextInEnd string
	//Строка выхода в другие комнаты, использовается для формирования сток команд "Идти" и "Осмотреться"
	TextExitRooms string
	//Общая строка для команды "Идти"
	StringPrintToMove string
	//Общая строка для команды "Осмотреться"
	StringPrintToLookAround string
	//Флаг, что комната закрыта и в нее нельзя пройти
	IsLocked bool
	//Строка для вывода при попытке зайти в закрытую комнату
	LockedString string
	//Сопастовление имен комнат и объектов комнат
	ExitRooms map[string]*Room
	//Слайс всех контейнеров, которые доступны в комнате
	Containers []*Container
	//Таргеты, которые доступны в комнате
	Targets map[string]*Target
}

type Item struct {
	//Название предмета
	Name string
	//Флаг, который проверяет можно ли взять предмет в инвентарь
	IsTake bool
	//Флаг, который проверяет можно ли надеть предмеи
	IsWearing bool
	//Функция предмета при применении
	OnApply func(p *Player) 
	//Функция предмета, если ее одевает игрок
	OnWearing func(p *Player) 
}

//------------База данных игры--------------//

var PlayerGame *Player
var CommandsGame map[string]func(player *Player, args string) string 
var ItemsGame map[string]*Item
var RoomsGame map[string]*Room

//------------------------------------------//

//---------Вспомогательные функции--------------//
//Функция проверки предмета в контейнерах комнаты
func checkItemInContainer(containers []*Container, item string) *Item {
	var itemObject *Item

		for _, container := range containers {
			for _, itemInContainer := range container.InternalItems {
				if itemInContainer.Name == item {
					itemObject = itemInContainer
					break
				}
			} 
		}
	return itemObject
}
//Функция удаления предмета из контейнера после команды "Взять"
func deleteItemInContainer(containers []*Container, item *Item) bool {
		for index, container := range containers {
			for index, itemInContainer := range container.InternalItems {
				if itemInContainer == item {
					container.InternalItems = append(container.InternalItems[:index], container.InternalItems[index+1:]...)
					return true
				}
			} 
			if len(container.InternalItems) == 0 {
				containers = append(containers[:index], containers[index+1:]...)
			}
		}
	return false
}
//Функция удаления контейнера из комнаты, если в контейнере кмонаты не осталось предмета
func deleteContainerInRoom(room *Room) bool {
	for index, container := range room.Containers {
		if len(container.InternalItems) == 0 {
				room.Containers = append(room.Containers[:index], room.Containers[index+1:]...)
				return true
			}
	}
	return false
}
//Функция формирования строки для команды "Осмотреться", которая собирает каждый контейнер и предметы в нем
func createStringPrintToLookAround(room *Room) string {
	var sliceStringContainer []string
	var stringContainer string
	if len(room.Containers) != 0 {
		for _, container := range room.Containers {
			var sliceItemName []string
			
			for _, item := range container.InternalItems {
				sliceItemName = append(sliceItemName, item.Name)
			}
			sliceStringContainer = append(sliceStringContainer, container.Name + strings.Join(sliceItemName, ", "))
		}
		stringContainer = strings.Join(sliceStringContainer, ", ")
		return stringContainer
	} else {
		return fmt.Sprintf("пустая %s", room.Name)
	}
}
//Функция обновления строкии для команды "Осмотреться"
func updateStringToLook(room *Room)  {
	room.StringPrintToLookAround = room.UniqueTextInStart + createStringPrintToLookAround(room) + room.UniqueTextInEnd + room.TextExitRooms

}

//------------------------------------------//



func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	reader := bufio.NewScanner(os.Stdin)
	var command string

	initGame()
	for {
		reader.Scan()
		command = reader.Text()
		fmt.Println(handleCommand(command))
	}
	
	
}

func initGame() {
	//--------- CREATE ITEMS-----------//

	tea := &Item{
		Name: "чай",
		IsTake: true,
		OnApply: nil,
	}

	key := &Item{
		Name: "ключи",
		IsTake: true,
		OnApply: func(p *Player)  {
			RoomsGame["улица"].IsLocked = false
		},
	}

	notes := &Item{
		Name: "конспекты",
		IsTake: true,
		OnApply: nil,
	}

	backpack := &Item{
		Name: "рюкзак",
		IsTake: true,
		IsWearing: true,
		OnWearing: func(p *Player) {
			p.IsInventoryAccess = true
			RoomsGame["кухня"].StringPrintToLookAround = strings.ReplaceAll(RoomsGame["кухня"].StringPrintToLookAround, "собрать рюкзак и ", "")
			RoomsGame["кухня"].UniqueTextInEnd = strings.ReplaceAll(RoomsGame["кухня"].UniqueTextInEnd, "собрать рюкзак и ", "")
		},
	}

	// Заполняем базу всех предметов игры

	ItemsGame = map[string]*Item{
		"чай": tea,
		"ключи": key,
		"рюкзак": backpack,
		"конспекты": notes,
	}

	//-------------------------------//

	//--------- CREATE CONTAINER-----------//

	kitchen_table := &Container{
		Name: "на столе: ",
		InternalItems: []*Item{
			tea,
		},
	}

	player_room_table := &Container{
		Name: "на столе: ",
		InternalItems: []*Item{
			key,
			notes,
		},
	}

	player_room_chair := &Container{
		Name: "на стуле: ",
		InternalItems: []*Item{
			backpack,
		},
	}

	//-------------------------------//

	//--------- CREATE TARGER-----------//

	targetDorInCoridor := &Target{
		Name: "дверь",
		ItemsApply: map[string]*Item{
			"ключи": key,
		},
		ApplyResult: map[*Item]string{
			key: "открыта",
		},
	}

	targetWardrobeInCoridor := &Target{
		Name: "шкаф",
		ItemsApply: map[string]*Item{},
	}
	//-------------------------------//


	//--------- CREATE ROOM-----------//

	coridor := &Room{
		Name: "коридор",
		Description: "ничего интересного",
		TextExitRooms: ". можно пройти - кухня, комната, улица",
		Targets: map[string]*Target{
			"дверь": targetDorInCoridor,
			"шкаф": targetWardrobeInCoridor,
		},
	}

	kitchen := &Room{
		Name: "кухня",
		UniqueTextInStart: "ты находишься на кухне, ",
		UniqueTextInEnd: ", надо собрать рюкзак и идти в универ",
		Description: "кухня, ничего интересного",
		TextExitRooms: ". можно пройти - коридор",
		IsLocked: false,
		Containers: []*Container{
			kitchen_table,
		},
	}

	playerRoom := &Room{
		Name: "комната",
		Description: "ты в своей комнате",
		TextExitRooms: ". можно пройти - коридор",
		IsLocked: false,
		Containers: []*Container{
			player_room_table,
			player_room_chair,
		},
	}

	street := &Room{
		Name: "улица",
		Description: "на улице весна",
		TextExitRooms: ". можно пройти - домой",
		IsLocked: true,
		LockedString: "дверь закрыта",
	}

	kitchen.ExitRooms = map[string]*Room{
		"коридор": coridor,
	}
	coridor.ExitRooms = map[string]*Room{
		"кухня": kitchen,
		"комната": playerRoom,
		"улица": street,
	}
	playerRoom.ExitRooms = map[string]*Room{
		"коридор": coridor,
	}
	street.ExitRooms = map[string]*Room{
		"домой": coridor,
	}
	kitchen.StringPrintToMove = kitchen.Description + kitchen.TextExitRooms
	coridor.StringPrintToMove = coridor.Description + coridor.TextExitRooms
	playerRoom.StringPrintToMove = playerRoom.Description + playerRoom.TextExitRooms
	street.StringPrintToMove = street.Description + street.TextExitRooms

	kitchen.StringPrintToLookAround = kitchen.UniqueTextInStart + createStringPrintToLookAround(kitchen) + kitchen.UniqueTextInEnd + kitchen.TextExitRooms
	coridor.StringPrintToLookAround = coridor.UniqueTextInStart + createStringPrintToLookAround(coridor) + coridor.UniqueTextInEnd + coridor.TextExitRooms
	playerRoom.StringPrintToLookAround = playerRoom.UniqueTextInStart + createStringPrintToLookAround(playerRoom) + playerRoom.UniqueTextInEnd + playerRoom.TextExitRooms
	street.StringPrintToLookAround = street.UniqueTextInStart + createStringPrintToLookAround(street) + street.UniqueTextInEnd + street.TextExitRooms


	RoomsGame = map[string]*Room{
		"кухня": kitchen,
		"комната": playerRoom,
		"коридор": coridor,
		"улица": street,
	}

	//-------------------------------//

	//--------- CREATE PLAYER-----------//

	player := &Player{
		IsInventoryAccess: false,
		Inventory: make(map[string]*Item),
		CurrentLocathion: kitchen,
	}

	PlayerGame = player

	//-------------------------------//

	//--------- CREATE COMMANDS-----------//

	putOn := func(player *Player, item string) string {
		item = strings.ReplaceAll(item, " ", "")

		itemObject := checkItemInContainer(player.CurrentLocathion.Containers, item)

		if itemObject == nil {
			return "нет такого"
		}
		
		if itemObject.IsWearing {
			itemObject.OnWearing(player)

			isDeleted := deleteItemInContainer(player.CurrentLocathion.Containers, itemObject)
		
			if isDeleted {
				deleteContainerInRoom(player.CurrentLocathion)
			}
			updateStringToLook(player.CurrentLocathion)
			

			return fmt.Sprintf("вы надели: %s", itemObject.Name)
		} else {
			return "предмет не надевается"
		}
	}

	toTake := func(player *Player, item string) string {
		item = strings.ReplaceAll(item, " ", "")

		

		itemObject := checkItemInContainer(player.CurrentLocathion.Containers, item)

		if itemObject == nil {
			return "нет такого"
		}
		
	if !player.IsInventoryAccess {
			return "некуда класть"
		}

		player.Inventory[itemObject.Name] = itemObject
		isDeleted := deleteItemInContainer(player.CurrentLocathion.Containers, itemObject)
		

		if isDeleted {
			deleteContainerInRoom(player.CurrentLocathion)
		}

		updateStringToLook(player.CurrentLocathion)

		return fmt.Sprintf("предмет добавлен в инвентарь: %s", itemObject.Name)
	}	

	apply := func(player *Player, args string) string {
		objects := strings.Fields(args)
		itemUse := objects[0]
		targetToApply := objects[1]
		var target *Target
		var item *Item
		
		if itemSearch, ok := player.Inventory[itemUse]; !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", itemUse)
		} else {
			item = itemSearch
		}

		if targetSearch, ok := player.CurrentLocathion.Targets[targetToApply]; !ok {
			return "такого нет"
		} else {
			target = targetSearch
		}

		

		if _, ok := target.ItemsApply[item.Name]; !ok {
			return "не к чему применить"
		} else {
			item.OnApply(player)
			return fmt.Sprintf("%s %s", target.Name, target.ApplyResult[item])
		}
	}

	move := func(player *Player, roomGo string) string {
		roomGo = strings.ReplaceAll(roomGo, " ", "")
		if nextRoom, ok := player.CurrentLocathion.ExitRooms[roomGo]; !ok {
			return fmt.Sprintf("нет пути в %s", roomGo)
		} else {
			if nextRoom.IsLocked {
				return nextRoom.LockedString
			}
			player.CurrentLocathion = nextRoom
			return nextRoom.StringPrintToMove
		}
	}

	toLookAround := func(player *Player, roomGo string) string {
		return player.CurrentLocathion.StringPrintToLookAround
	}

	CommandsGame = map[string]func(player *Player, args string) string{
		"идти": move,
		"взять": toTake,
		"надеть": putOn,
		"применить": apply,
		"осмотреться": toLookAround,

	}
	//-------------------------------//

}

func handleCommand(command string) string {
	var funcActhion func(player *Player, roomGo string) string
	commandSlice := strings.Fields(command)
	if len(commandSlice) == 0 {
		return "введите команду"
	}
	if funcCmd, ok := CommandsGame[commandSlice[0]]; !ok {
		return "неизвестная команда"
	} else {
		funcActhion = funcCmd
	}
	stringArgs := ""
	for i := 1; i < len(commandSlice); i++ {
		stringArgs += commandSlice[i] + " "
	}
	
	
	return funcActhion(PlayerGame, stringArgs)
}
