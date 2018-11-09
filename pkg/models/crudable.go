package models

// CRUDable defines the crud methods
type CRUDable interface {
	Create(*User) error
	ReadOne() error
	ReadAll(string, *User, int) (interface{}, error)
	Update() error
	Delete() error
}
