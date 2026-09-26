import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {downloadExportMutationOptions, exportKeys, requestExportMutationOptions} from './dataExport'
import {downloadBlob} from '@/helpers/downloadBlob'
const sdk = vi.hoisted(() => ({
	userExportRequest: vi.fn(),
	userExportDownload: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/downloadBlob', () => ({downloadBlob: vi.fn()}))

beforeEach(() => {
	vi.mocked(downloadBlob).mockClear()
	URL.createObjectURL = vi.fn(() => 'blob:export')
})

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

it('saves the downloaded export under the zip file name', async () => {
	const client = new QueryClient()
	sdk.userExportDownload.mockResolvedValue({data: new Blob(['PK'])})
	await client.getMutationCache().build(client, downloadExportMutationOptions()).execute('password')
	expect(sdk.userExportDownload).toHaveBeenCalledWith({
		body: {password: 'password'},
		parseAs: 'blob',
	})
	expect(downloadBlob).toHaveBeenCalledWith('blob:export', 'vikunja-export.zip')
})

it('rejects a download response that is not a file', async () => {
	const client = new QueryClient()
	sdk.userExportDownload.mockResolvedValue({data: {message: 'wrong password'}})
	await expect(client.getMutationCache().build(client, downloadExportMutationOptions()).execute('password'))
		.rejects.toThrow('Export response was not a blob')
	expect(downloadBlob).not.toHaveBeenCalled()
})
