import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {totpQuery, totpKeys, enableTotpMutationOptions, disableTotpMutationOptions} from './totp'
const sdk = vi.hoisted(() => ({
	totpGet: vi.fn(),
	totpEnable: vi.fn(),
	totpDisable: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('treats absent enrollment as disabled without retrying', async () => {
	const client = new QueryClient({defaultOptions: {queries: {retry: false}}})
	sdk.totpGet.mockRejectedValue({code: 1016})
	expect(await client.fetchQuery(totpQuery())).toEqual({enabled: false})
})
it('drops the enrollment secret and QR data after enabling', async () => {
	const client = new QueryClient()
	client.setQueryData(totpKeys.current, {
		enabled: false,
		secret: 'SHAREDSECRET',
	})
	client.setQueryData(totpKeys.qr, new Blob())
	sdk.totpEnable.mockResolvedValue({data: {}})
	await client.getMutationCache().build(client, enableTotpMutationOptions()).execute('123456')
	expect(sdk.totpEnable).toHaveBeenCalledWith({body: {passcode: '123456'}})
	expect(client.getQueryData(totpKeys.current)).toEqual({enabled: true})
	expect(client.getQueryData(totpKeys.qr)).toBeUndefined()
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
