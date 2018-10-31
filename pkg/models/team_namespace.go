package models

// TeamNamespace defines the relationship between a Team and a Namespace
type TeamNamespace struct {
	ID          int64     `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID      int64     `xorm:"int(11) not null INDEX" json:"team_id" param:"team"`
	NamespaceID int64     `xorm:"int(11) not null INDEX" json:"namespace_id" param:"namespace"`
	Right       TeamRight `xorm:"int(11) INDEX" json:"right"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (TeamNamespace) TableName() string {
	return "team_namespaces"
}
