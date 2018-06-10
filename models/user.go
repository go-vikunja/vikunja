package models

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
)

// UserLogin Object to recive user credentials in JSON format
type UserLogin struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

// User holds information about an user
type User struct {
	ID       int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Name     string `xorm:"varchar(250)" json:"name"`
	Username string `xorm:"varchar(250) not null unique" json:"username"`
	Password string `xorm:"varchar(250) not null" json:"password"`
	Email    string `xorm:"varchar(250)" json:"email"`
	IsAdmin  bool   `xorm:"tinyint(1) not null" json:"isAdmin"`
	Created  int64  `xorm:"created" json:"created"`
	Updated  int64  `xorm:"updated" json:"updated"`
}

// TableName returns the table name for users
func (User) TableName() string {
	return "users"
}

// GetUserByID gets informations about a user by its ID
func GetUserByID(id int64) (user User, exists bool, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returing one even if the ID is 0
	if id == 0 {
		return User{}, false, nil
	}

	return GetUser(User{ID: id})
}

// GetUser gets a user object
func GetUser(user User) (userOut User, exists bool, err error) {
	userOut = user
	exists, err = x.Get(&userOut)

	if !exists {
		return User{}, false, ErrUserDoesNotExist{}
	}

	return userOut, exists, err
}

// CheckUserCredentials checks user credentials
func CheckUserCredentials(u *UserLogin) (User, error) {

	// Check if the user exists
	user, exists, err := GetUser(User{Username: u.Username})
	if err != nil {
		return User{}, err
	}

	if !exists {
		return User{}, ErrUserDoesNotExist{}
	}

	// Check the users password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))

	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetCurrentUser returns the current user based on its jwt token
func GetCurrentUser(c echo.Context) (user User, err error) {
	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	userID, ok := claims["id"].(float64)
	if !ok {
		return user, ErrCouldNotGetUserID{}
	}
	user = User{
		ID:       int64(userID),
		Name:     claims["name"].(string),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}

	return
}
