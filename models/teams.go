package models

// Team holds a team object
type Team struct {
	ID          int64   `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Name        string  `xorm:"varchar(250) not null" json:"name"`
	Description string  `xorm:"varchar(250)" json:"description"`
	Rights      []int64 `xorm:"varchar(250)" json:"rights"`
	CreatedByID int64   `xorm:"int(11) not null" json:"-"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CreatedBy User `json:"created_by"`
}

// TableName makes beautiful table names
func (Team) TableName() string {
	return "teams"
}

// TeamMember defines the relationship between a user and a team
type TeamMember struct {
	ID     int64 `xorm:"int(11) autoincr not null unique pk"`
	TeamID int64 `xorm:"int(11) autoincr not null"`
	UserID int64 `xorm:"int(11) autoincr not null"`

	Created int64 `xorm:"created"`
	Updated int64 `xorm:"updated"`
}

// TableName makes beautiful table names
func (TeamMember) TableName() string {
	return "team_members"
}

// TeamNamespaces defines the relationship between a Team and a Namespace
type TeamNamespace struct {
	ID          int64 `xorm:"int(11) autoincr not null unique pk"`
	TeamID      int64 `xorm:"int(11) autoincr not null"`
	NamespaceID int64 `xorm:"int(11) autoincr not null"`

	Created int64 `xorm:"created"`
	Updated int64 `xorm:"updated"`
}

// TableName makes beautiful table names
func (TeamNamespace) TableName() string {
	return "team_namespaces"
}

// TeamList defines the relation between a team and a list
type TeamList struct {
	ID     int64 `xorm:"int(11) autoincr not null unique pk"`
	TeamID int64 `xorm:"int(11) autoincr not null"`
	ListID int64 `xorm:"int(11) autoincr not null"`

	Created int64 `xorm:"created"`
	Updated int64 `xorm:"updated"`
}

// TableName makes beautiful table names
func (TeamList) TableName() string {
	return "team_list"
}

func GetAllTeamsByNamespaceID(id int64) (teams []*Team, err error) {
	err = x.Table("teams").
		Join("INNER", "team_namespaces", "teams.id = team_id").
		Where("teams.namespace_id = ?", id).
		Find(teams)

	return
}
