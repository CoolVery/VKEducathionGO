package models

type Room struct {
	Name string
	Description string
	StringPrintToMove string
	IsLocked bool
	ExitRooms map[string]*Room
	Containers []*Container
	Targets map[string]*Target
}