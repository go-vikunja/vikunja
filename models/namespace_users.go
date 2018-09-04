package models

// NamespaceUser represents a namespace <-> user relation
type NamespaceUser struct {
	ID          int64     `xorm:"int(11) autoincr not null unique pk" json:"id" param:"namespace"`
	UserID      int64     `xorm:"int(11) not null" json:"user_id" param:"user"`
	NamespaceID int64     `xorm:"int(11) not null" json:"namespace_id" param:"namespace"`
	Right       UserRight `xorm:"int(11)" json:"right"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName is the table name for NamespaceUser
func (NamespaceUser) TableName() string {
	return "users_namespace"
}
