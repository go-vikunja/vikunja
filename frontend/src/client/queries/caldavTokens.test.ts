import {it, expect, vi} from 'vitest'
import {defineComponent, h} from 'vue'
import {mount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {
	caldavTokenKeys,
	caldavTokensQuery,
	createCaldavTokenMutationOptions,
	deleteCaldavTokenMutationOptions,
	useCreateCaldavTokenMutation,
} from './caldavTokens'
const sdk = vi.hoisted(() => ({
	caldavTokensCreate: vi.fn(),
	caldavTokensList: vi.fn(),
	caldavTokensDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('returns the one-time secret without putting it in the shared list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.all, [{id: 1}])
	sdk.caldavTokensCreate.mockResolvedValue({data: {id: 2, token: 'one-time-secret'}})
	const created = await client.getMutationCache().build(client, createCaldavTokenMutationOptions()).execute(undefined)
	expect(created.token).toBe('one-time-secret')
	expect(client.getQueryData(caldavTokenKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(caldavTokenKeys.all)?.isInvalidated).toBe(true)
})

it('does not keep the one-time secret in the mutation cache', async () => {
	const client = new QueryClient()
	sdk.caldavTokensCreate.mockResolvedValue({data: {id: 2, token: 'one-time-secret'}})
	let mutation!: ReturnType<typeof useCreateCaldavTokenMutation>
	const wrapper = mount(defineComponent({
		setup() {
			mutation = useCreateCaldavTokenMutation()
			return () => h('div')
		},
	}), {
		global: {
			plugins: [[VueQueryPlugin, {queryClient: client}]],
		},
	})

	expect((await mutation.mutateAsync()).token).toBe('one-time-secret')
	await new Promise(resolve => setTimeout(resolve))

	expect(client.getMutationCache().getAll()).toEqual([])
	wrapper.unmount()
})

it('loads the tokens from the first page only', async () => {
	const client = new QueryClient()
	sdk.caldavTokensList.mockResolvedValue({
		data: {
			items: [{id: 1}, {id: 2}],
			total_pages: 2,
		},
	})
	expect(await client.fetchQuery(caldavTokensQuery())).toEqual([{id: 1}, {id: 2}])
	expect(sdk.caldavTokensList).toHaveBeenCalledTimes(1)
	expect(sdk.caldavTokensList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 1}}))
})

it('treats a missing token list as empty', async () => {
	const client = new QueryClient()
	sdk.caldavTokensList.mockResolvedValue({data: {items: null}})
	expect(await client.fetchQuery(caldavTokensQuery())).toEqual([])
})

it('removes a deleted token from an existing cache and marks the list stale', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.all, [{id: 1}, {id: 2}])
	sdk.caldavTokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteCaldavTokenMutationOptions()).execute(2)
	expect(sdk.caldavTokensDelete).toHaveBeenCalledWith({path: {id: 2}})
	expect(client.getQueryData(caldavTokenKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(caldavTokenKeys.all)?.isInvalidated).toBe(true)
})
