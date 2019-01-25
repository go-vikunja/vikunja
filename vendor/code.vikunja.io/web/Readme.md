# Vikunja Web Handler

[![License: LGPL v3](https://img.shields.io/badge/License-LGPL%20v3-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/code.vikunja.io/web)](https://goreportcard.com/report/code.vikunja.io/web)

> When I started Vikunja, I started like everyone else, by writing a bunch of functions to do the logic and then a bunch of
handler functions to parse the request data and call the implemented functions to do the logic and eventually return a dataset.
After I implemented some functions, I've decided to save me a lot of hassle and put most of that "parse the request and call a 
processing function"-logic to a general interface to facilitate development and not having to have a lot of similar code all over the place.

This webhandler was built to be used in a REST-API, it takes and returns JSON, but can also be used in combination with own other handler
implementations thus leading to much flexibility.

## Features

* Easy to use
* Built for REST-APIs
* Beautiful error handling built in
* Manages rights
* Pluggable authentication mechanisms

## Table of contents

* [Installation](#installation)
* [Todos](#todos)
* [CRUDable](#crudable)
* [Rights](#rights)
* [Handler Config](#handler-config)
  * [Auth](#auth)
  * [Logging](#logging)
  * [Full Example](#full-example)
* [Preprocessing](#preprocessing)
  * [Pagination](#pagination)
  * [Search](#search)
* [Standard web handler](#standard-web-handler)
* [Errors](#errors)
* [URL param binder](#how-the-url-param-binder-works)

### TODOs

* [ ] Description of web.HTTPError
* [ ] Rights methods should return errors (I know, this will break a lot of existing stuff)
* [ ] Improve docs

## Installation

Using the web handler in your application is pretty straight forward, simply run `go get -u code.vikunja.io/web` and start using it. 

In order to use the common web handler, the struct must implement the `web.CRUDable` and `web.Rights` interface.

## CRUDable

This interface defines methods to Create/Read/ReadAll/Update/Delete something. It is defined as followed:

```go
type CRUDable interface {
	Create(Auth) error
	ReadOne() error
	ReadAll(string, Auth, int) (interface{}, error)
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

## Rights

This interface defines methods to check for rights on structs. They accept an `Auth`-element as parameter and return a `bool`.

The interface is defined as followed:

```go
type Rights interface {
	IsAdmin(Auth) bool
	CanWrite(Auth) bool
	CanRead(Auth) bool
	CanDelete(Auth) bool
	CanUpdate(Auth) bool
	CanCreate(Auth) bool
}
```

When using the standard web handler, all methods except `CanRead()` are called before their `CRUD` counterparts. `CanRead()`
is called after `ReadOne()` was invoked as this would otherwise mean getting an object from the db to check if the user has the
right to see it and then getting it again if thats the case. Calling the function afterwards means we only have to get the
object once.

## Handler Config

The handler has some options which you can (and need to) configure.

#### Auth

`Auth` is an interface with some methods to decouple the action of getting the current user from the web handler.
The function defined via `Auths` should return a struct which implements the `Auth` interface.

To define the thing which gets the appropriate auth object, you need to call a middleware like so (After all auth middlewares were called):

#### Logging

You can provide your own instance of `logger.Logger` (using [this package](https://github.com/op/go-logging)) to the handler.
It will use this instance to log errors which are not better specified or things like users trying to do something they're
not allowed to do and so on.

#### Full Example

```go
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        c.Set("AuthProvider", &web.Auths{
            AuthObject: func(echo.Context) (web.Auth, error) {
                return models.GetCurrentUser(c) // Your functions
            },
        })
        c.Set("LoggingProvider", &log.Log)
        return next(c)
    }
})
```

## Preprocessing

### Pagination

When using the `ReadAll`-method, the third parameter contains the requested page. Your function should return only the number of results
corresponding to that page. The number of items per page is definied in the config as `service.pagecount` (Get it with `viper.GetInt("service.pagecount")`).

These can be calculated in combination with a helper function, `getLimitFromPageIndex(pageIndex)` which returns
SQL-needed `limit` (max-length) and `offset` parameters. You can feed this function directly into xorm's `Limit`-Function like so:

```go
lists := []List{}
err := x.Limit(getLimitFromPageIndex(pageIndex)).Find(&lists)
```

### Search

When using the `ReadAll`-method, the first parameter is a search term which should be used to search items of your struct. You define the critera.

Users can then pass the `?s=something` parameter to the url to search.

As the logic for "give me everything" and "give me everything where the name contains 'something'" is mostly the same, we made the decision to design 
the function like this, in order to keep the places with mostly the same logic as few as possible. Also just adding `?s=query` to the url one already 
knows and uses is a lot more convenient.

## Standard web handler

You can define routes for the standard web handler like so:

`models.List` needs to implement `web.CRUDable` and `web.Rights`.

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

The handler will take care of everything like parsing the request, checking rights, pretty-print errors and return appropriate responses.

## Errors

Error types with their messages and http-codes should be implemented by you somewhere in your application and then returned by 
the appropriate function when an error occures. If the error type implements `HTTPError`, the server returns a user-friendly 
error message when this error occours. This means it returns a good HTTP status code, a message, and an error code. The error 
code should be unique across all error codes and can be used on the client to show a localized error message or do other stuff 
based on the exact error the server returns. That way the client won't have to "guess" that the error message remains the same 
over multiple versions of your application.

An `HTTPError` is defined as follows:

```go
type HTTPError struct {
	HTTPCode int    `json:"-"` // Can be any valid HTTP status code, I'd reccomend to use the constants of the http package.
	Code     int    `json:"code"` // Must be a uniqe int identifier for this specific error. I'd reccomend defining a constant for this.
	Message  string `json:"message"` // A user-readable message what went wrong.
}
```

## How the url param binder works

The binder binds all values inside the url to their respective fields in a struct. Those fields need to have a tag
"param" with the name of the url placeholder which must be the same as in routes.

Whenever one of the standard CRUD methods is invoked, this binder is called, which enables one handler method
to handle all kinds of different urls with different parameters.
