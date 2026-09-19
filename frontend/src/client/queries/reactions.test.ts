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
} from './tasks'
import {setReactionMutationOptions} from './reactions'

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

it('adds only the current user and preserves other reactions', async () => {
	client.setQueryData(taskKeys.detail(1), normalizeTask({
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
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({reactions: {
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
	client.setQueryData(taskKeys.detail(1), normalizeTask({
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
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({reactions: {'👍': [{id: 2}]}})
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
