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
		for _, container := range containers {
			for index, itemInContainer := range container.InternalItems {
				if itemInContainer == item {
					container.InternalItems = append(container.InternalItems[:index], container.InternalItems[index+1:]...)
					return true
				}
			} 
		}
	return false
}

// func createStringPrintToMove(room *Room) string {
// 	var stringPrint string
// 	stringPrint += room.Description

// 	var stringNextRoom string
// 	if len(room.Containers) != 0 {
// 		for _, container := range room.Containers {
// 		stringContainer += container.Name
// 			for _, item := range container.InternalItems {
// 				stringContainer += " " + item.Name + ","
// 			}
// 		}
// 	}



// 	stringPrint += stringContainer

	

// 	return stringPrint
// }
//------------------------------------------//



func main() {
	/*
		в этой функции можно ничего не писать,
		но тогда у вас не будет работать через go run main.go
		очень круто будет сделать построчный ввод команд тут, хотя это и не требуется по заданию
	*/
	initGame()
	fmt.Println(handleCommand("идти комната"))
	fmt.Println(handleCommand("идти коридор"))
	fmt.Println(handleCommand("применить ключи дверь"))
	fmt.Println(handleCommand("идти комната"))
	fmt.Println(handleCommand("взять ключи"))
	fmt.Println(handleCommand("надеть рюкзак"))
	fmt.Println(handleCommand("взять ключи"))
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
		Name: "На столе:",
		InternalItems: []*Item{
			tea,
		},
	}

	player_room_table := &Container{
		Name: "На столе:",
		InternalItems: []*Item{
			key,
			notes,
		},
	}

	player_room_chair := &Container{
		Name: "На стуле:",
		InternalItems: []*Item{
			backpack,
		},
	}

	//-------------------------------//

	//--------- CREATE TARGER-----------//

	targetDorInCoridor := &Target{
		Name: "дверь",
		ItemsApply: map[string]*Item{
			"ключ": key,
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
		Description: "ничего интересного.",
		Targets: map[string]*Target{
			"дверь": targetDorInCoridor,
			"шкаф": targetWardrobeInCoridor,
		},
	}

	kitchen := &Room{
		Name: "кухня",
		Description: "кухня, ничего интересного.",
		IsLocked: false,
		Containers: []*Container{
			kitchen_table,
		},
	}

	playerRoom := &Room{
		Name: "комната",
		Description: "ты в своей комнате.",
		IsLocked: false,
		Containers: []*Container{
			player_room_chair,
			player_room_table,
		},
	}

	street := &Room{
		Name: "улица",
		Description: "на улице весна.",
		IsLocked: true,
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
	kitchen.StringPrintToMove = kitchen.Description + " можно пройти - коридор"
	coridor.StringPrintToMove = coridor.Description + " можно пройти - кухня, комната, улица"
	playerRoom.StringPrintToMove = playerRoom.Description + " можно пройти - коридор"
	street.StringPrintToMove = street.Description + " можно пройти - домой"



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
		
		itemObject := checkItemInContainer(player.CurrentLocathion.Containers, item)

		if itemObject == nil {
			return "нет такого"
		}
		
		if itemObject.IsWearing {
			itemObject.OnWearing(player)
			return fmt.Sprintf("вы надели: %s", itemObject.Name)
		} else {
			return "предмет не надевается"
		}
	}

	toTake := func(player *Player, item string) string {

		if !player.IsInventoryAccess {
			return "некуда класть"
		}

		itemObject := checkItemInContainer(player.CurrentLocathion.Containers, item)

		if itemObject == nil {
			return "нет такого"
		}
		
		player.Inventory[itemObject.Name] = itemObject
		isDeleted := deleteItemInContainer(player.CurrentLocathion.Containers, itemObject)
		
		if !isDeleted {
			panic("Что-то пошло не так")
		}

		return fmt.Sprintf("предмет добавлен в инвентарь: %s", itemObject.Name)
	}	

	apply := func(player *Player, args string) string {
		objects := strings.Fields(args)
		itemUse := objects[0]
		targetToApply := objects[1]
		var target *Target
		var item *Item
		
		if target, ok := player.CurrentLocathion.Targets[targetToApply]; !ok {
			return "такого нет"
		} else {
			target = target
		}

		if item, ok := player.Inventory[itemUse]; !ok {
			return fmt.Sprintf("нет предмета в инвентаре - %s", itemUse)
		} else {
			item = item
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
			player.CurrentLocathion = nextRoom
			return fmt.Sprintf("%s", nextRoom.StringPrintToMove)
		}
	}

	CommandsGame = map[string]func(player *Player, args string) string{
		"идти": move,
		"взять": toTake,
		"надеть": putOn,
		"применить": apply,

	}
	//-------------------------------//

}

func handleCommand(command string) string {
	
	commandSlice := strings.Fields(command)
	funcActhion := CommandsGame[commandSlice[0]]
	stringArgs := ""
	for i := 1; i < len(commandSlice); i++ {
		stringArgs += commandSlice[i] + " "
	}
	
	
	return funcActhion(PlayerGame, stringArgs)
}
