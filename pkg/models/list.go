package models

// List represents a list of tasks
type List struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	Title       string `xorm:"varchar(250)" json:"title"`
	Description string `xorm:"varchar(1000)" json:"description"`
	OwnerID     int64  `xorm:"int(11) INDEX" json:"-"`
	NamespaceID int64  `xorm:"int(11) INDEX" json:"-" param:"namespace"`

	Owner User        `xorm:"-" json:"owner"`
	Tasks []*ListTask `xorm:"-" json:"tasks"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
func (l *List) ReadAll(u *User) (interface{}, error) {
	lists := []*List{}
	fullUser, err := GetUserByID(u.ID)
	if err != nil {
		return lists, err
	}

	// Gets all Lists where the user is either owner or in a team which has access to the list
	// Or in a team which has namespace read access
	err = x.Select("l.*").
		Table("list").
		Alias("l").
		Join("INNER", []string{"namespaces", "n"}, "l.namespace_id = n.id").
		Join("LEFT", []string{"team_namespaces", "tn"}, "tn.namespace_id = n.id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Where("tm.user_id = ?", fullUser.ID).
		Or("tm2.user_id = ?", fullUser.ID).
		Or("l.owner_id = ?", fullUser.ID).
		Or("ul.user_id = ?", fullUser.ID).
		Or("un.user_id = ?", fullUser.ID).
		GroupBy("l.id").
		Find(&lists)

	// Add more list details
	AddListDetails(lists)

	return lists, err
}

// ReadOne gets one list by its ID
func (l *List) ReadOne() (err error) {
	err = l.GetSimpleByID()
	if err != nil {
		return err
	}

	// Get list tasks
	l.Tasks, err = GetTasksByListID(l.ID)
	if err != nil {
		return err
	}

	// Get list owner
	l.Owner, err = GetUserByID(l.OwnerID)
	return
}

// GetSimpleByID gets a list with only the basic items, aka no tasks or user objects. Returns an error if the list does not exist.
func (l *List) GetSimpleByID() (err error) {
	if l.ID < 1 {
		return ErrListDoesNotExist{ID: l.ID}
	}

	// We need to re-init our list object, because otherwise xorm creates a "where for every item in that list object,
	// leading to not finding anything if the id is good, but for example the title is different.
	id := l.ID
	*l = List{}
	exists, err := x.Where("id = ?", id).Get(l)
	if err != nil {
		return
	}

	if !exists {
		return ErrListDoesNotExist{ID: l.ID}
	}

	return
}

// AddListDetails adds owner user objects and list tasks to all lists in the slice
func AddListDetails(lists []*List) (err error) {
	var listIDs []int64
	var ownerIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
		ownerIDs = append(ownerIDs, l.OwnerID)
	}

	// Get all tasks
	ts := []*ListTask{}
	err = x.In("list_id", listIDs).Find(&ts)
	if err != nil {
		return
	}

	// Get all list owners
	owners := []*User{}
	err = x.In("id", ownerIDs).Find(&owners)
	if err != nil {
		return
	}

	// Build it all into the lists slice
	for in, list := range lists {
		// Owner
		for _, owner := range owners {
			if list.OwnerID == owner.ID {
				lists[in].Owner = *owner
				break
			}
		}

		// Tasks
		for _, task := range ts {
			if task.ListID == list.ID {
				lists[in].Tasks = append(lists[in].Tasks, task)
			}
		}
	}

	return
}
