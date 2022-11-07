import type {IBucket} from '@/modelTypes/IBucket'
import type {IList} from '@/modelTypes/IList'

const key = 'collapsedBuckets'

export type CollapsedBuckets = {[id: IBucket['id']]: boolean}

function getAllState() {
	const saved = localStorage.getItem(key)
	return saved === null
		? {}
		: JSON.parse(saved)
}

export const saveCollapsedBucketState = (
	listId: IList['id'],
	collapsedBuckets: CollapsedBuckets,
) => {
	const state = getAllState()
	state[listId] = collapsedBuckets
	for (const bucketId in state[listId]) {
		if (!state[listId][bucketId]) {
			delete state[listId][bucketId]
		}
	}
	localStorage.setItem(key, JSON.stringify(state))
}

export function getCollapsedBucketState(listId : IList['id']) {
	const state = getAllState()
	return typeof state[listId] !== 'undefined'
		? state[listId]
		: {}
}
