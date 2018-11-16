package models

// TeamList defines the relation between a team and a list
type TeamList struct {
	ID     int64     `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID int64     `xorm:"int(11) not null INDEX" json:"team_id" param:"team"`
	ListID int64     `xorm:"int(11) not null INDEX" json:"list_id" param:"list"`
	Right  TeamRight `xorm:"int(11) INDEX" json:"right" valid:"length(0|2)"`

	Created int64 `xorm:"created" json:"created" valid:"range(0|0)"`
	Updated int64 `xorm:"updated" json:"updated" valid:"range(0|0)"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (TeamList) TableName() string {
	return "team_list"
}

// TeamWithRight represents a team, combined with rights.
type TeamWithRight struct {
	Team  `xorm:"extends"`
	Right TeamRight `json:"right"`
}
