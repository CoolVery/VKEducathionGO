package models

type Target struct {
	Name string
	ItemsApply map[string]*Item
	ApplyResult map[*Item]string
}