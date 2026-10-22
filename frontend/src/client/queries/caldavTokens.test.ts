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

function tokenPage(page: number, ids: number[], total: number) {
	return {
		items: ids.map(id => ({id})),
		page,
		per_page: 2,
		total,
		total_pages: Math.ceil(total / 2),
	}
}

it('returns the one-time secret without putting it in the shared list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.list(1), tokenPage(1, [1], 1))
	sdk.caldavTokensCreate.mockResolvedValue({data: {id: 2, token: 'one-time-secret'}})
	const created = await client.getMutationCache().build(client, createCaldavTokenMutationOptions()).execute(undefined)
	expect(created.token).toBe('one-time-secret')
	expect(client.getQueryData(caldavTokenKeys.list(1))).toEqual(tokenPage(1, [1], 1))
	expect(client.getQueryState(caldavTokenKeys.list(1))?.isInvalidated).toBe(true)
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

it('requests only the asked-for page', async () => {
	const client = new QueryClient()
	sdk.caldavTokensList.mockResolvedValue({data: tokenPage(2, [3], 3)})
	expect(await client.fetchQuery(caldavTokensQuery(2))).toEqual(tokenPage(2, [3], 3))
	expect(sdk.caldavTokensList).toHaveBeenCalledTimes(1)
	expect(sdk.caldavTokensList).toHaveBeenCalledWith(expect.objectContaining({
		query: {page: 2},
	}))
})

it('treats a missing token list as empty', async () => {
	const client = new QueryClient()
	sdk.caldavTokensList.mockResolvedValue({data: {items: null}})
	expect((await client.fetchQuery(caldavTokensQuery(1))).items).toEqual([])
})

it('removes a deleted token and rewrites the totals on every cached page', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.list(1), tokenPage(1, [1, 2], 3))
	client.setQueryData(caldavTokenKeys.list(2), tokenPage(2, [3], 3))
	sdk.caldavTokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteCaldavTokenMutationOptions()).execute(2)
	expect(sdk.caldavTokensDelete).toHaveBeenCalledWith({path: {id: 2}})
	expect(client.getQueryData(caldavTokenKeys.list(1))).toEqual(tokenPage(1, [1], 2))
	expect(client.getQueryData(caldavTokenKeys.list(2))).toEqual(tokenPage(2, [3], 2))
	expect(client.getQueryState(caldavTokenKeys.list(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(caldavTokenKeys.list(2))?.isInvalidated).toBe(true)
})
