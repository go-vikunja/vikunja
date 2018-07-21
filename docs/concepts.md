# Achitectural concepts

Vikunja was built with a maximum flexibility in mind while developing. To achive this, I built a set of easy-to-use
functions and respective web handlers, all represented through interfaces.

## CRUDable

This interface defines methods to Create/Read/ReadAll/Update/Delete something. In order to use the common web
handler, the struct mus implement this and the rights interface.

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