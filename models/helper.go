package models

import (
	"github.com/labstack/echo"
	"strconv"
)

// GetIntURLParam is a helper method which returns an int from an url param
func GetIntURLParam(param string, c echo.Context) (intParam int64, err error) {

	id := c.Param(param)
	if id != "" {
		intParam, err = strconv.ParseInt(id, 10, 64)
	}

	return intParam, err
}

// GetByID gets an object by its ID
func GetByID(id int64, result interface{}) (err error) {
	exists, err := x.ID(id).Get(result)
	if err != nil {
		return err
	}

	if !exists {
		return ErrListDoesNotExist{}
	}

	return
}
