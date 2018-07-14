package models

// Team holds a team object
type Team struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Name        string `xorm:"varchar(250) not null" json:"name"`
	Description string `xorm:"varchar(250)" json:"description"`
	CreatedByID int64  `xorm:"int(11) not null" json:"-"`

	CreatedBy *User   `xorm:"-" json:"created_by"`
	Members   []*User `xorm:"-" json:"members"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (Team) TableName() string {
	return "teams"
}

func (t *Team) AfterLoad() {
	// Get the owner
	*t.CreatedBy, _, _ = GetUserByID(t.CreatedByID)
}

// TeamMember defines the relationship between a user and a team
type TeamMember struct {
	ID     int64 `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID int64 `xorm:"int(11) not null" json:"team_id"`
	UserID int64 `xorm:"int(11) not null" json:"user_id"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`
}

// TableName makes beautiful table names
func (TeamMember) TableName() string {
	return "team_members"
}

// TeamNamespace defines the relationship between a Team and a Namespace
type TeamNamespace struct {
	ID          int64   `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID      int64   `xorm:"int(11) not null" json:"team_id"`
	NamespaceID int64   `xorm:"int(11) not null" json:"namespace_id"`
	Rights      []int64 `xorm:"varchar(250)" json:"rights"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`
}

// TableName makes beautiful table names
func (TeamNamespace) TableName() string {
	return "team_namespaces"
}

// TeamList defines the relation between a team and a list
type TeamList struct {
	ID     int64   `xorm:"int(11) autoincr not null unique pk" json:"id"`
	TeamID int64   `xorm:"int(11) not null" json:"team_id"`
	ListID int64   `xorm:"int(11) not null" json:"list_id"`
	Rights []int64 `xorm:"varchar(250)" json:"rights"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`
}

// TableName makes beautiful table names
func (TeamList) TableName() string {
	return "team_list"
}

// GetAllTeamsByNamespaceID returns all teams for a namespace
func GetAllTeamsByNamespaceID(id int64) (teams []*Team, err error) {
	err = x.Table("teams").
		Join("INNER", "team_namespaces", "teams.id = team_id").
		Where("teams.namespace_id = ?", id).
		Find(teams)

	return
}

// Empty empties a struct. Because we heavily use pointers, the old values remain in the struct.
// If you then update by not providing evrything, you have i.e. the old description still in the
// newly created team,  but you didn't provided one.
func (t *Team) Empty() {
	t.ID = 0
	t.CreatedByID = 0
	t.CreatedBy = &User{}
	t.Name = ""
	t.Description = ""
	t.Members = []*User{}
}