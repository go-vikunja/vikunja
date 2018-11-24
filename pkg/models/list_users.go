package models

// ListUser represents a list <-> user relation
type ListUser struct {
	ID     int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	UserID int64     `xorm:"int(11) not null INDEX" json:"user_id" param:"user"`
	ListID int64     `xorm:"int(11) not null INDEX" json:"list_id" param:"list"`
	Right  UserRight `xorm:"int(11) INDEX" json:"right" valid:"length(0|2)"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for ListUser
func (ListUser) TableName() string {
	return "users_list"
}

// UserWithRight represents a user in combination with the right it can have on a list/namespace
type UserWithRight struct {
	User  `xorm:"extends"`
	Right UserRight `json:"right"`
}
