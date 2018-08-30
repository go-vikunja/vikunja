package models

// Delete deletes a namespace
func (n *Namespace) Delete() (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(n.ID).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their tasks
	lists, err := GetListsByNamespaceID(n.ID)
	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list tasks
	for _, list := range lists {
		listIDs = append(listIDs, list.ID)
	}

	// Delete tasks
	_, err = x.In("list_id", listIDs).Delete(&ListTask{})
	if err != nil {
		return
	}

	// Delete the lists
	_, err = x.In("id", listIDs).Delete(&List{})
	if err != nil {
		return
	}

	return
}
