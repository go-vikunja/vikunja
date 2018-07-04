package models

func DeleteNamespaceByID(namespaceID int64, doer *User) (err error) {

	// Check if the namespace exists
	namespace, err := GetNamespaceByID(namespaceID)
	if err != nil {
		return
	}

	// Check if the user is namespace admin
	err = doer.IsNamespaceAdmin(&namespace)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(namespaceID).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their items
	lists, err := GetListsByNamespaceID(namespaceID)
	var listIDs []int64
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
