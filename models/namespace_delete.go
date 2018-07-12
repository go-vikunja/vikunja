package models

// Delete deletes a namespace
func (n *Namespace) Delete(id int64) (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(id)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(id).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their items
	lists, err := GetListsByNamespaceID(id)
	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list items
	for _, list := range lists {
		listIDs = append(listIDs, list.ID)
	}

	// Delete items
	_, err = x.In("list_id", listIDs).Delete(&ListItem{})
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
