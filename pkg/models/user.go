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
	Username string `xorm:"varchar(250) not null unique" json:"username" valid:"length(5|250)"`
	Password string `xorm:"varchar(250) not null" json:"-"`
	Email    string `xorm:"varchar(250)" json:"email" valid:"email,length(0|250)"`
	IsActive bool   `json:"-"`

	PasswordResetToken string `xorm:"varchar(450)" json:"-"`
	EmailConfirmToken  string `xorm:"varchar(450)" json:"-"`

	Created int64 `xorm:"created" json:"created" valid:"range(0|0)"`
	Updated int64 `xorm:"updated" json:"updated" valid:"range(0|0)"`
}

// TableName returns the table name for users
func (User) TableName() string {
	return "users"
}

// APIUserPassword represents a user object without timestamps and a json password field.
type APIUserPassword struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// APIFormat formats an API User into a normal user struct
func (apiUser *APIUserPassword) APIFormat() User {
	return User{
		ID:       apiUser.ID,
		Username: apiUser.Username,
		Password: apiUser.Password,
		Email:    apiUser.Email,
	}
}

// GetUserByID gets informations about a user by its ID
func GetUserByID(id int64) (user User, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returing one even if the ID is 0
	if id < 1 {
		return User{}, ErrUserDoesNotExist{}
	}

	return GetUser(User{ID: id})
}

// GetUser gets a user object
func GetUser(user User) (userOut User, err error) {
	userOut = user
	exists, err := x.Get(&userOut)

	if !exists {
		return User{}, ErrUserDoesNotExist{UserID: user.ID}
	}

	return userOut, err
}

// CheckUserCredentials checks user credentials
func CheckUserCredentials(u *UserLogin) (User, error) {
	// Check if we have any credentials
	if u.Password == "" || u.Username == "" {
		return User{}, ErrNoUsernamePassword{}
	}

	// Check if the user exists
	user, err := GetUser(User{Username: u.Username})
	if err != nil {
		return User{}, err
	}

	// User is invalid if it needs to verify its email address
	if !user.IsActive {
		return User{}, ErrEmailNotConfirmed{UserID: user.ID}
	}

	// Check the users password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return User{}, ErrWrongUsernameOrPassword{}
		}
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
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}

	return
}
