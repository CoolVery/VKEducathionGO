package main

import (
	. "firstHomework/models"
	"fmt"
	"strings"
)

//------------База данных игры--------------//

var PlayerGame *Player
var CommandsGame map[string]func(player *Player, args string) string 
var ItemsGame map[string]*Item
var RoomsGame map[string]*Room

//------------------------------------------//

//---------Вспомогательные функции--------------//

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

func deleteContainerInRoom(room *Room) bool {
	for index, container := range room.Containers {
		if len(container.InternalItems) == 0 {
				room.Containers = append(room.Containers[:index], room.Containers[index+1:]...)
				return true
			}
	}
	return false
}

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
	initGame()
	fmt.Println(handleCommand("осмотреться"))            // 1  осмотр на кухне
fmt.Println(handleCommand("идти коридор"))           // 2  переход в коридор
fmt.Println(handleCommand("идти комната"))           // 3  переход в комнату
fmt.Println(handleCommand("осмотреться"))            // 4  осмотр комнаты
fmt.Println(handleCommand("надеть рюкзак"))          // 5  надеть рюкзак
fmt.Println(handleCommand("взять ключи"))            // 6  взять ключи
fmt.Println(handleCommand("взять конспекты"))        // 7  взять конспекты
fmt.Println(handleCommand("идти коридор"))           // 8  обратно в коридор
fmt.Println(handleCommand("применить ключи дверь"))  // 9  открыть дверь
fmt.Println(handleCommand("идти улица"))             // 10 выйти на улицу
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

		if !player.IsInventoryAccess {
			return "некуда класть"
		}

		itemObject := checkItemInContainer(player.CurrentLocathion.Containers, item)

		if itemObject == nil {
			return "нет такого"
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
		
		if targetSearch, ok := player.CurrentLocathion.Targets[targetToApply]; !ok {
			return "такого нет"
		} else {
			target = targetSearch
		}

		if itemSearch, ok := player.Inventory[itemUse]; !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", itemUse)
		} else {
			item = itemSearch
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
			return fmt.Sprintf("%s", nextRoom.StringPrintToMove)
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
