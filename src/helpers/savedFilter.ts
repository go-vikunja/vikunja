import ListModel from '@/models/list'

export function getSavedFilterIdFromListId(listId: ListModel['id']) {
	let filterId = listId * -1 - 1
	// FilterIds from listIds are always positive
	if (filterId < 0) {
		filterId = 0
	}
	return filterId
}