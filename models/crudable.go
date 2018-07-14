package models

// CRUDable defines the crud methods
type CRUDable interface {
	Create(*User, int64) error
	ReadOne(int64) error
	ReadAll(*User) (interface{}, error)
	Update(int64) error
	Delete(int64) error
}
