package models

type Player struct {
	Inventory map[string]*Item
	IsInventoryAccess bool
	CurrentLocathion *Room
}