package models

type Game struct {
	Player *Player
	Rooms map[string]*Room
}