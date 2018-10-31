package models

// CRUDable defines the crud methods
type CRUDable interface {
	Create(*User) error
	ReadOne() error
	ReadAll(*User) (interface{}, error)
	Update() error
	Delete() error
}
