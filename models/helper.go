package models

import (
	"fmt"
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

func GetAllByUser(user *User, result interface{}) (err error) {
	fmt.Println(result)
	err = x.Where("owner_id = ", user.ID).Find(result)
	return
}
