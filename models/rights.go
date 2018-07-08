package models

type Rights interface {
	IsAdmin(*User) bool
}
