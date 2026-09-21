import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {exportKeys, requestExportMutationOptions} from './dataExport'
const sdk = vi.hoisted(() => ({userExportRequest: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('marks existing export metadata stale without inventing a completed export', async () => {
	const client = new QueryClient()
	client.setQueryData(exportKeys.status, null)
	const before = client.getQueryState(exportKeys.status)?.dataUpdateCount
	sdk.userExportRequest.mockResolvedValue({data: {}})
	await client.getMutationCache().build(client, requestExportMutationOptions()).execute('password')
	expect(sdk.userExportRequest).toHaveBeenCalledWith({body: {password: 'password'}})
	expect(client.getQueryState(exportKeys.status)?.dataUpdateCount).toBe(before)
	expect(client.getQueryState(exportKeys.status)?.isInvalidated).toBe(true)
})
