import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {totpQuery, totpKeys, disableTotpMutationOptions} from './totp'
const sdk = vi.hoisted(() => ({totpGet: vi.fn(), totpDisable: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('treats absent enrollment as disabled without retrying', async () => {
	const client = new QueryClient({defaultOptions: {queries: {retry: false}}})
	sdk.totpGet.mockRejectedValue({code: 1016})
	expect(await client.fetchQuery(totpQuery())).toEqual({enabled: false})
})
it('drops enrollment secrets and QR data after disabling', async () => {
	const client = new QueryClient()
	client.setQueryData(totpKeys.current, {enabled: true, secret: 'old'})
	client.setQueryData(totpKeys.qr, new Blob())
	sdk.totpDisable.mockResolvedValue({data: {}})
	await client.getMutationCache().build(client, disableTotpMutationOptions()).execute('password')
	expect(client.getQueryData(totpKeys.current)).toEqual({enabled: false})
	expect(client.getQueryData(totpKeys.qr)).toBeUndefined()
})
