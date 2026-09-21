import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {caldavTokenKeys, createCaldavTokenMutationOptions} from './caldavTokens'
const sdk = vi.hoisted(() => ({caldavTokensCreate: vi.fn()}))
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
