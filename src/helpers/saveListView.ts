// Save the current list view to local storage

import type { IList } from '@/modelTypes/IList'

type ListView = Record<IList['id'], string>

const DEFAULT_LIST_VIEW = 'list.list' as const

// We use local storage and not a store here to make it persistent across reloads.
export const saveListView = (listId: IList['id'], routeName: string) => {
	if (routeName.includes('settings.')) {
		return
	}

	if (!listId) {
		return
	}

	const savedListView = localStorage.getItem('listView')
	let savedListViewJson: ListView | false = false
	if (savedListView !== null) {
		savedListViewJson = JSON.parse(savedListView) as ListView
	}

	let listView: ListView = {}
	if (savedListViewJson) {
		listView = savedListViewJson
	}

	listView[listId] = routeName
	localStorage.setItem('listView', JSON.stringify(listView))
}

export const getListView = (listId: IList['id']) => {
	// Remove old stored settings
	const savedListView = localStorage.getItem('listView')
	if (savedListView !== null && savedListView.startsWith('list.')) {
		localStorage.removeItem('listView')
	}

	if (!savedListView) {
		return DEFAULT_LIST_VIEW
	}

	const savedListViewJson: ListView = JSON.parse(savedListView)

	if (!savedListViewJson[listId]) {
		return DEFAULT_LIST_VIEW
	}

	return savedListViewJson[listId]
}