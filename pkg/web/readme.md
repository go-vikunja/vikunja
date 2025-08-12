# Vikunja Web Handler

[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/code.vikunja.io/web)](https://goreportcard.com/report/code.vikunja.io/web)

> When I started Vikunja, I started like everyone else, by writing a bunch of functions to do the logic and then a bunch of
handler functions to parse the request data and call the implemented functions to do the logic and eventually return a dataset.
After I implemented some functions, I've decided to save me a lot of hassle and put most of that "parse the request and call a 
processing function"-logic to a general interface to facilitate development and not having to have a lot of similar code all over the place.

This webhandler was built to be used in a REST-API, it takes and returns JSON, but can also be used in combination with own 
other handler implementations, enabling a lot of flexibility while developing.

## Features

* Easy to use
* Built for REST-APIs
* Beautiful error handling built in
* Manages permissions
* Pluggable authentication mechanisms

## Table of contents

* [Installation](#installation)
* [Todos](#todos)
* [DB Sessions](#db-sessions)
* [CRUDable](#crudable)
* [Permissions](#permissions)
* [Handler Config](#handler-config)
  * [Auth](#auth)
  * [Logging](#logging)
  * [Full Example](#full-example)
* [Preprocessing](#preprocessing)
  * [Pagination](#pagination)
  * [Search](#search)
* [Standard web handler](#defining-routes-using-the-standard-web-handler)
* [Errors](#errors)
* [URL param binder](#how-the-url-param-binder-works)

### TODOs

* [x] Improve docs/Merge with the ones of Vikunja
* [x] Description of web.HTTPError
* [x] Permissions methods should return errors (I know, this will break a lot of existing stuff)
* [ ] optional Before- and after-{load|update|create} methods which do some preprocessing/after processing like making human-readable names from automatically up counting consts
* [ ] "Magic": Check if a passed struct implements Crudable methods and use a general (user defined) function if not

## Installation

Using the web handler in your application is pretty straight forward, simply run `go get -u code.vikunja.io/web` and start using it. 

In order to use the common web handler, the struct must implement the `web.CRUDable` and `web.Permissions` interface.

To learn how to use the handler, take a look at the [handler config](#handler-config) [defining routes](#defining-routes-using-the-standard-web-handler)

## DB Sessions

Each request runs in its own db session.
This ensures each operation is one atomic entity without any side effects for concurrent requests happening at the same time.

The session is started at the beginning of the request, rolled back in case of any errors and committed if no errors occur.
The permissions methods get the same session (for the same request) as the actual crud methods.

See [`SessionFactory`](#sessionfactory) for docs about how to configure it.

## CRUDable

This interface defines methods to Create/Read/ReadAll/Update/Delete something. It is defined as followed:

```go
type CRUDable interface {
	Create(*xorm.Session, Auth) error
	ReadOne(*xorm.Session, Auth) error
	ReadAll(s *xorm.Session, auth Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error)
	Update(*xorm.Session, Auth) error
	Delete(*xorm.Session, Auth) error
}
```

Each of these methods gets called on an instance of a struct like so:

```go
func (l *List) ReadOne() (err error) {
	*l, err = GetListByID(l.ID)
	return
}
```

In that case, it takes the `ID` saved in the struct instance, gets the full list object and fills the original object with it.
(See [parambinder](#how-the-url-param-binder-works) to understand where that `ID` is coming from in that specific case).

All functions should behave like this, if they create or update something, the struct instance they are called on should 
contain the created/updated struct instance. The only exception is `ReadAll()` which returns an interface. 
Usually this method returns a slice of results because you cannot make an array of a set type (If you know a 
way to do this, don't hesitate to [drop me a message](https://vikunja.io/en/contact/)).

## Permissions

This interface defines methods to check for permissions on structs. They accept an `Auth`-element as parameter and return a `bool` and `error`.

The `error` is handled [as usual](#errors).

The interface is defined as followed:

```go
type Permissions interface {
	CanRead(*xorm.Session, Auth) (bool, int, error) // The int is the max permission the user has for this entity.
	CanDelete(*xorm.Session, Auth) (bool, error)
	CanUpdate(*xorm.Session, Auth) (bool, error)
	CanCreate(*xorm.Session, Auth) (bool, error)
}
```

When using the standard web handler, all methods are called before their `CRUD` counterparts.
Use pointers for methods like `CanRead()` to get the base data of the model first, then check the permission and then add additional data.

The `CanRead` method should also return the max permission a user has on this entity.
This number will be returned in the `x-max-permission` header to enable user interfaces to show/hide UI elements based on the permission the user has.

## Handler Config

The handler has some options which you can (and need to) configure.

#### Auth

`Auth` is an interface with some methods to decouple the action of getting the current user from the web handler.
The function defined via `Auths` should return a struct which implements the `Auth` interface.

To define the thing which gets the appropriate auth object, you need to call a middleware like so (After all auth middlewares were called):

#### Logging

You can provide your own instance of `slog.Logger` from Go's standard library to the handler.
It will use this instance to log errors which are not better specified or things like users trying to do something they're
not allowed to do and so on.

#### MaxItemsPerPage

Contains the maximum number of items per page.
If the client requests more items than this, the number of items requested is set to this value.

See [pagination](#pagination) for more.

#### SessionFactory

To create a new session for each request, you need to call the `SetSessionFactory` method before any web request.
It has the following signature:

```go
func SetSessionFactory(sessionFactory func() *xorm.Session)
```

The closure will be called for every request.

#### Full Example

```go
handler.SetAuthProvider(&web.Auths{
    AuthObject: func(echo.Context) (web.Auth, error) {
        return models.GetCurrentUser(c) // Your functions
    },
})
handler.SetLoggingProvider(&log.Log)
handler.SetSessionFactory(x.NewSession)
```

## Preprocessing

### Pagination

The `ReadAll`-method has a number of parameters:

```go
ReadAll(auth Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfItems int64, err error)
```

The third parameter contains the requested page, the fourth parameter contains the number of items per page.
You should calculate the limits accordingly.

If the number of items per page are not set by the client, the web handler will pass the maximum number of items per page instead.
This makes items per page optional for clients.
Take a look at [the config section](#handler-config) for information on how to set that value.

You need to return a number of things:

* The result itself, usually a slice
* The number of items you return in `result`. Most of the time, this is just `len(result)`. You need to return this value to make the clients aware if they requested a number of items > max items per page.
* The total number of items available. We use the total number of items here and not the number of pages so implementations don't have to calculate the page count. The total number is returned to the client; it can be used to build client-side pagination or similar.
* An error.

The number of items and the total number of pages available will be returned in the `x-pagination-total-pages` and `x-pagination-result-count` response headers.
_You should put this in your api documentation._

### Search

When using the `ReadAll`-method, the first parameter is a search term which should be used to search items of your struct. 
You define the criteria inside of that function.

Users can then pass the `?s=something` parameter to the url to search, _thats something you should put in your api documentation_.

As the logic for "give me everything" and "give me everything where the name contains 'something'" is mostly the same, we made 
the decision to design the function like this, in order to keep the places with mostly the same logic as few as possible. 
Also just adding `?s=query` to the url one already knows and uses is a lot more convenient.

## Defining routes using the standard web handler

You can define routes for the standard web handler like so:

`models.List` needs to implement `web.CRUDable` and `web.Permissions`.

```go
listHandler := &crud.WebHandler{
    EmptyStruct: func() crud.CObject {
        return &models.List{}
    },
}
a.GET("/lists", listHandler.ReadAllWeb)
a.GET("/lists/:list", listHandler.ReadOneWeb)
a.POST("/lists/:list", listHandler.UpdateWeb)
a.DELETE("/lists/:list", listHandler.DeleteWeb)
a.PUT("/namespaces/:namespace/lists", listHandler.CreateWeb)
```

The handler will take care of everything like parsing the request, checking permissions, pretty-print errors and return appropriate responses.

## Errors

Error types with their messages and http-codes should be implemented by you somewhere in your application and then returned by 
the appropriate function when an error occurs. If the error type implements `HTTPError`, the server returns a user-friendly
error message when this error occurs. This means it returns a good HTTP status code, a message, and an error code. The error
code should be unique across all error codes and can be used on the client to show a localized error message or do other stuff 
based on the exact error the server returns. That way the client won't have to "guess" that the error message remains the same 
over multiple versions of your application.

An `HTTPError` is defined as follows:

```go
type HTTPError struct {
    HTTPCode int    `json:"-"` // Can be any valid HTTP status code, I'd recommend to use the constants of the http package.
    Code     int    `json:"code"` // Must be a unique int identifier for this specific error. I'd recommend defining a constant for this.
	Message  string `json:"message"` // A user-readable message what went wrong.
}
```

You can learn more about how exactly custom error types are created in the [vikunja docs](https://vikunja.io/docs/custom-errors/).

## How the url param binder works

The binder binds all values inside the url to their respective fields in a struct. Those fields need to have a tag
`param` with the name of the url placeholder which must be the same as in routes.

Whenever one of the standard CRUD methods is invoked, this binder is called, which enables one handler method
to handle all kinds of different urls with different parameters.
