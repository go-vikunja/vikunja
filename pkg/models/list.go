package models

import "sort"

// List represents a list of tasks
type List struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"list"`
	Title       string `xorm:"varchar(250)" json:"title" valid:"required,runelength(5|250)"`
	Description string `xorm:"varchar(1000)" json:"description" valid:"runelength(0|1000)"`
	OwnerID     int64  `xorm:"int(11) INDEX" json:"-"`
	NamespaceID int64  `xorm:"int(11) INDEX" json:"-" param:"namespace"`

	Owner User        `xorm:"-" json:"owner"`
	Tasks []*ListTask `xorm:"-" json:"tasks"`

	Created int64 `xorm:"created" json:"created" valid:"range(0|0)"`
	Updated int64 `xorm:"updated" json:"updated" valid:"range(0|0)"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

// ReadAll gets all lists a user has access to
// @Summary Get all lists a user has access to
// @Description Returns all lists a user has access to.
// @tags list
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search lists by title."
// @Security ApiKeyAuth
// @Success 200 {array} models.List "The lists"
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists [get]
func (l *List) ReadAll(search string, u *User, page int) (interface{}, error) {
	lists, err := getRawListsForUser(search, u, page)
	if err != nil {
		return nil, err
	}

	// Add more list details
	AddListDetails(lists)

	return lists, err
}

// ReadOne gets one list by its ID
// @Summary Gets one list
// @Description Returns a list by its ID.
// @tags list
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.List "The list"
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [get]
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

// Gets the lists only, without any tasks or so
func getRawListsForUser(search string, u *User, page int) (lists []*List, err error) {
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
		Limit(getLimitFromPageIndex(page)).
		Where("l.title LIKE ?", "%"+search+"%").
		Find(&lists)

	return lists, err
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

// ReadAll gets all tasks for a user
// @Summary Get tasks
// @Description Returns all tasks on any list the user has access to.
// @tags task
// @Accept json
// @Produce json
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search tasks by task text."
// @Security ApiKeyAuth
// @Success 200 {array} models.List "The tasks"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks [get]
func (lt *ListTask) ReadAll(search string, u *User, page int) (interface{}, error) {
	return GetTasksByUser(search, u, page)
}

//GetTasksByUser returns all tasks for a user
func GetTasksByUser(search string, u *User, page int) (tasks []*ListTask, err error) {
	// Get all lists
	lists, err := getRawListsForUser("", u, page)
	if err != nil {
		return nil, err
	}

	// Get all list IDs and get the tasks
	var listIDs []int64
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	// Then return all tasks for that lists
	if err := x.In("list_id", listIDs).Where("text LIKE ?", "%"+search+"%").Find(&tasks); err != nil {
		return nil, err
	}

	// Sort it by due date
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDateUnix > tasks[j].DueDateUnix
	})

	return tasks, err
}
