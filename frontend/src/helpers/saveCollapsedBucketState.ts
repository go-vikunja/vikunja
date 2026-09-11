import type {IBucket} from '@/modelTypes/IBucket'

const key = 'collapsedBuckets'

export type CollapsedBuckets = {[id: IBucket['id']]: boolean}

function getAllState() {
	const saved = localStorage.getItem(key)
	return saved === null
		? {}
		: JSON.parse(saved)
}

export const saveCollapsedBucketState = (
	projectId: number,
	collapsedBuckets: CollapsedBuckets,
) => {
	const state = getAllState()
	state[projectId] = collapsedBuckets
	for (const bucketId of Object.keys(state[projectId] ?? {})) {
		if (!state[projectId][bucketId]) {
			delete state[projectId][bucketId]
		}
	}
	localStorage.setItem(key, JSON.stringify(state))
}

export function getCollapsedBucketState(projectId: number) {
	const state = getAllState()
	return typeof state[projectId] !== 'undefined'
		? state[projectId]
		: {}
}
