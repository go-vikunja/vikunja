package models

import (
	"github.com/labstack/echo"
	"strconv"
)

func GetIntURLParam(param string, c echo.Context) (intParam int64, err error) {

	id := c.Param(param)
	if id != "" {
		intParam, err = strconv.ParseInt(id, 10, 64)
	}

	return intParam, err
}
