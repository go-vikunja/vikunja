
export const saveListView = (listId, routeName) => {
	const savedListViewJson = JSON.parse(localStorage.getItem('listView'))

	let listView = {}
	if(savedListViewJson) {
		listView = savedListViewJson
	}

	listView[listId] = routeName
	localStorage.setItem('listView', JSON.stringify(listView))
}

export const getListView = listId => {
	// Remove old stored settings
	const savedListView = localStorage.getItem('listView')
	if(savedListView !== null && savedListView.startsWith('list.')) {
		localStorage.removeItem('listView')
	}

	console.log('saved list view state', savedListView)

	if (!savedListView) {
		return 'list.list'
	}

	const savedListViewJson = JSON.parse(savedListView)

	if(!savedListViewJson[listId]) {
		return 'list.list'
	}

	return savedListViewJson[listId]
}