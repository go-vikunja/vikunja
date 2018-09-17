package models

// ListUser represents a list <-> user relation
type ListUser struct {
	ID     int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	UserID int64     `xorm:"int(11) not null" json:"user_id" param:"user"`
	ListID int64     `xorm:"int(11) not null" json:"list_id" param:"list"`
	Right  UserRight `xorm:"int(11)" json:"right"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for ListUser
func (ListUser) TableName() string {
	return "users_list"
}

type userWithRight struct {
	User  `xorm:"extends"`
	Right UserRight `json:"right"`
}
