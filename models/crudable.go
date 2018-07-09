package models

type CRUDable interface {
	Create(*User) (error)
	ReadOne(int64) error
	ReadAll(*User) (interface{}, error)
	Update(int64, *User) error
	Delete()
}
