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

// GetListByID returns a list by its ID
func GetListByID(id int64) (list List, err error) {
	if id < 1 {
		return list, ErrListDoesNotExist{ID: id}
	}

	exists, err := x.ID(id).Get(&list)
	if err != nil {
		return list, err
	}

	if !exists {
		return list, ErrListDoesNotExist{ID: id}
	}

	// Get list tasks
	list.Tasks, err = GetTasksByListID(list.ID)
	if err != nil {
		return list, err
	}

	// Get list owner
	list.Owner, err = GetUserByID(list.OwnerID)

	return list, err
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
func (l *List) ReadAll(user *User) (interface{}, error) {
	lists := []*List{}
	fullUser, err := GetUserByID(user.ID)
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
	addListDetails(lists)

	return lists, err
}

// ReadOne gets one list by its ID
func (l *List) ReadOne() (err error) {
	*l, err = GetListByID(l.ID)
	return
}

// Adds owner user objects and list tasks to all lists in the slice
func addListDetails(lists []*List) (err error) {
	var listIDs []int64
	var ownerIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
		ownerIDs = append(ownerIDs, l.OwnerID)
	}

	// Get all tasks
	tasks := []*ListTask{}
	err = x.In("list_id", listIDs).Find(&tasks)
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
		for _, task := range tasks {
			if task.ListID == list.ID {
				lists[in].Tasks = append(lists[in].Tasks, task)
			}
		}
	}

	return
}
