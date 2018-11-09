# Architectural concepts

Vikunja was built with a maximum flexibility in mind while developing. To achive this, I built a set of easy-to-use
functions and respective web handlers, all represented through interfaces.

## CRUDable

This interface defines methods to Create/Read/ReadAll/Update/Delete something. In order to use the common web
handler, the struct must implement this and the `Rights` interface.

The interface is defined as followed:

```go
type CRUDable interface {
	Create(*User) error
	ReadOne() error
	ReadAll(*User, int) (interface{}, error)
	Update() error
	Delete() error
}
```

Each of these methods is called on an instance of a struct like so:

```go
func (l *List) ReadOne() (err error) {
	*l, err = GetListByID(l.ID)
	return
}
```

In that case, it takes the `ID` saved in the struct instance, gets the full list object and fills the original object with it.
(See parambinder to understand where that `ID` is coming from).

All functions should behave like this, if they create or update something, they should return the created/updated struct
instance. The only exception is `ReadAll()` which returns an interface. Usually this is an array, because, well you cannot
make an array of a set type (If you know a way to do this, don't hesitate to drop me a message).

### Pagination

When using the `ReadAll`-method, the second parameter contains the requested page. Your function should return only the number of results
corresponding to that page. The number of items per page is definied in the config as `service.pagecount` (Get it with `viper.GetInt("service.pagecount")`).

These can be calculated in combination with a helper function, `getLimitFromPageIndex(pageIndex)` which returns
SQL-needed `limit` (max-length) and `offset` parameters. You can feed this function directly into xorm's `Limit`-Function like so:

```go
lists := []List{}
err := x.Limit(getLimitFromPageIndex(pageIndex)).Find(&lists)
```

## Rights

This interface defines methods to check for rights on structs. They accept a `User` as parameter and usually return a `bool`.

The interface is defined as followed:

```go
type Rights interface {
	IsAdmin(*User) bool
	CanWrite(*User) bool
	CanRead(*User) bool
	CanDelete(*User) bool
	CanUpdate(*User) bool
	CanCreate(*User) bool
}
```

When using the standard web handler, all methods except `CanRead()` are called before their `CRUD` counterparts. `CanRead()`
is called after `ReadOne()` was invoked as this would otherwise mean getting an object from the db to check if the user has the
right to see it and then getting it again if thats the case. Calling the function afterwards means we only have to get the
object once. 

## Standard web handler

## Errors

Error types with their messages and http-codes are set in `models/error.go`. If the error type implements `HTTPError`, the server returns a user-friendly error message when this error occours. This means it returns a good HTTP status code, a message, and an error code. The error code should be unique across all error codes and can be used on the client to show a localized error message or do other stuff based on the exact error the server returns. That way the client won't have to "guess" that the error message remains the same over multiple versions of Vikunja.

An `HTTPError` is defined as follows:

```go
type HTTPError struct {
	HTTPCode int    `json:"-"` // Can be any valid HTTP status code, I'd reccomend to use the constants of the http package.
	Code     int    `json:"code"` // Must be a uniqe int identifier for this specific error. I'd reccomend defining a constant for this.
	Message  string `json:"message"` // A user-readable message what went wrong.
}
```

