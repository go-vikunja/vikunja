package models

type CRUDable interface {
	Create()
	ReadOne(int64) error
	ReadAll(*User) (interface{}, error)
	Update()
	Delete()
}
