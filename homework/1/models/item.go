package models

type Item struct {
	//Название предмета
	Name string
	//Флаг, который проверяет можно ли взять предмет в инвентарь
	IsTake bool
	IsWearing bool
	//Функция предмета
	OnApply func(p *Player) 
	OnWearing func(p *Player) 
}