
export function getSavedFilterIdFromListId(listId) {
	let filterId = listId * -1 - 1
	// FilterIds from listIds are always positive
	if (filterId < 0) {
		filterId = 0
	}
	return filterId
}