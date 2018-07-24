package models

// TeamList defines the relation between a team and a list
type TeamList struct {
	ID     int64     `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID int64     `xorm:"int(11) not null" json:"team_id" param:"team"`
	ListID int64     `xorm:"int(11) not null" json:"list_id" param:"list"`
	Right  TeamRight `xorm:"int(11)" json:"right"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (TeamList) TableName() string {
	return "team_list"
}
