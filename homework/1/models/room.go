package models

type Room struct {
	Name string
	Description string
	UniqueTextInStart string
	UniqueTextInEnd string
	TextExitRooms string
	StringPrintToMove string
	StringPrintToLookAround string
	IsLocked bool
	LockedString string
	ExitRooms map[string]*Room
	Containers []*Container
	Targets map[string]*Target
}