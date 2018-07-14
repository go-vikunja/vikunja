package models

// CRUDable defines the crud methods
type CRUDable interface {
	Create(*User, int64) error
	ReadOne(int64) error
	ReadAll(*User) (interface{}, error)
	Update(int64) error
	Delete(int64) error

	// This method is needed, because old values would otherwise remain in the struct.
	// TODO find a way of not needing an extra function
	Empty()
}
