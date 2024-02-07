import type {IBucket} from '@/modelTypes/IBucket'
import type {IProject} from '@/modelTypes/IProject'

const key = 'collapsedBuckets'

export type CollapsedBuckets = {[id: IBucket['id']]: boolean}

function getAllState() {
	const saved = localStorage.getItem(key)
	return saved === null
		? {}
		: JSON.parse(saved)
}

export const saveCollapsedBucketState = (
	projectId: IProject['id'],
	collapsedBuckets: CollapsedBuckets,
) => {
	const state = getAllState()
	state[projectId] = collapsedBuckets
	for (const bucketId in state[projectId]) {
		if (!state[projectId][bucketId]) {
			delete state[projectId][bucketId]
		}
	}
	localStorage.setItem(key, JSON.stringify(state))
}

export function getCollapsedBucketState(projectId : IProject['id']) {
	const state = getAllState()
	return typeof state[projectId] !== 'undefined'
		? state[projectId]
		: {}
}
