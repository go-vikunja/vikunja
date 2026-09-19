import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	normalizeTask,
	taskKeys,
	type TaskResponse,
} from './tasks'
import {setReactionMutationOptions} from './reactions'
import {error} from '@/message'

const sdk = vi.hoisted(() => ({
	reactionsCreate: vi.fn(),
	reactionsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
let client: QueryClient
beforeEach(() => {
	vi.clearAllMocks()
	client = new QueryClient()
})

function expandedDetail(id: number) {
	return taskKeys.detail(id, ['reactions'])
}

function cachedReactions(id: number) {
	return client.getQueryData<TaskResponse>(expandedDetail(id))?.reactions ?? {}
}

it('adds only the current user and preserves other reactions', async () => {
	client.setQueryData(expandedDetail(1), normalizeTask({
		id: 1,
		reactions: {
			'👍': [{id: 2}],
			'🎉': [{id: 3}],
		},
	}))
	sdk.reactionsCreate.mockResolvedValue({data: {
		value: '👍',
		user: {
			id: 1,
			name: 'Server Authored',
		},
	}})
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: false,
		user: {
			id: 1,
			name: 'Stale Local',
		},
	})
	expect(sdk.reactionsCreate).toHaveBeenCalledWith({
		path: {
			entitykind: 'tasks',
			entityid: 1,
		},
		body: {value: '👍'},
	})
	expect(client.getQueryData(expandedDetail(1))).toMatchObject({reactions: {
		'👍': [
			{id: 2},
			{
				id: 1,
				name: 'Server Authored',
			},
		],
		'🎉': [{id: 3}],
	}})
})

it('removes only the caller and leaves unmounted details absent', async () => {
	client.setQueryData(expandedDetail(1), normalizeTask({
		id: 1,
		reactions: {'👍': [{id: 1}, {id: 2}]},
	}))
	sdk.reactionsDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: true,
		user: {id: 1},
	})
	expect(sdk.reactionsDelete).toHaveBeenCalledWith({
		path: {
			entitykind: 'tasks',
			entityid: 1,
		},
		body: {value: '👍'},
	})
	expect(client.getQueryData(expandedDetail(1))).toMatchObject({reactions: {'👍': [{id: 2}]}})
	const count = client.getQueryCache().getAll().length
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 99,
		value: '👍',
		remove: true,
		user: {id: 1},
	})
	expect(client.getQueryCache().getAll()).toHaveLength(count)
})

it('keeps the emoji order stable when toggling an existing reaction', async () => {
	client.setQueryData(expandedDetail(1), normalizeTask({
		id: 1,
		reactions: {
			'🎉': [{id: 3}],
			'👍': [{id: 1}, {id: 2}],
			'❤️': [{id: 4}],
		},
	}))
	sdk.reactionsDelete.mockResolvedValue({data: undefined})
	sdk.reactionsCreate.mockResolvedValue({data: {value: '👍'}})
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: true,
		user: {id: 1},
	})
	expect(Object.keys(cachedReactions(1))).toEqual(['🎉', '👍', '❤️'])
	expect(cachedReactions(1)['👍']).toEqual([{id: 2}])
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: false,
		user: {id: 1},
	})
	expect(Object.keys(cachedReactions(1))).toEqual(['🎉', '👍', '❤️'])
	expect(cachedReactions(1)['👍']).toEqual([{id: 2}, {id: 1}])
})

it('patches only the copies that expanded reactions', async () => {
	const task = normalizeTask({
		id: 1,
		reactions: {'👍': [{id: 2}]},
	})
	client.setQueryData(expandedDetail(1), task)
	client.setQueryData(taskKeys.detail(1), task)
	client.setQueryData(taskKeys.allList({project: 1}), [task])
	sdk.reactionsCreate.mockResolvedValue({data: {
		value: '👍',
		user: {id: 1},
	}})
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: false,
		user: {id: 1},
	})
	expect(cachedReactions(1)).toEqual({'👍': [{id: 2}, {id: 1}]})
	expect(client.getQueryData(taskKeys.detail(1))).toBe(task)
	expect(client.getQueryData<TaskResponse[]>(taskKeys.allList({project: 1}))?.[0]).toBe(task)
})

it('leaves task caches untouched for a comment reaction', async () => {
	const task = normalizeTask({
		id: 1,
		reactions: {'👍': [{id: 2}]},
	})
	client.setQueryData(taskKeys.detail(1), task)
	client.setQueryData(taskKeys.allList({project: 1}), [task])
	sdk.reactionsCreate.mockResolvedValue({data: {
		value: '👍',
		user: {id: 1},
	}})
	await client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'comments',
		id: 1,
		value: '👍',
		remove: false,
		user: {id: 1},
	})
	expect(sdk.reactionsCreate).toHaveBeenCalledWith({
		path: {
			entitykind: 'comments',
			entityid: 1,
		},
		body: {value: '👍'},
	})
	expect(client.getQueryData(taskKeys.detail(1))).toBe(task)
	expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(false)
	expect(client.getQueryState(taskKeys.allList({project: 1}))?.isInvalidated).toBe(false)
})

it('keeps the cache and toasts once when the request fails', async () => {
	const task = normalizeTask({
		id: 1,
		reactions: {'👍': [{id: 2}]},
	})
	client.setQueryData(expandedDetail(1), task)
	sdk.reactionsCreate.mockRejectedValue(new Error('nope'))
	await expect(client.getMutationCache().build(client, setReactionMutationOptions()).execute({
		kind: 'tasks',
		id: 1,
		value: '👍',
		remove: false,
		user: {id: 1},
	})).rejects.toThrow('nope')
	expect(client.getQueryData(expandedDetail(1))).toBe(task)
	expect(error).toHaveBeenCalledTimes(1)
})
