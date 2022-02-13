import ListModel from '@/models/list'

const key = 'collapsedBuckets'

const getAllState = () => {
	const saved = localStorage.getItem(key)
	if (saved === null) {
		return {}
	}

	return JSON.parse(saved)
}

export const saveCollapsedBucketState = (listId: ListModel['id'], collapsedBuckets) => {
	const state = getAllState()
	state[listId] = collapsedBuckets
	for (const bucketId in state[listId]) {
		if (!state[listId][bucketId]) {
			delete state[listId][bucketId]
		}
	}
	localStorage.setItem(key, JSON.stringify(state))
}

export const getCollapsedBucketState = (listId : ListModel['id']) => {
	const state = getAllState()
	if (typeof state[listId] !== 'undefined') {
		return state[listId]
	}

	return {}
}
