import {describe, it, expect, beforeEach, vi} from 'vitest'

import {fetchAttachmentBlob, fetchAttachmentUrl, releaseAttachmentUrl} from './attachments'
import {queryClient} from '@/client/queryClient'
import {attachmentKeys} from '@/client/queries/attachments'

const {getBlobUrl} = vi.hoisted(() => ({getBlobUrl: vi.fn()}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number, type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/client/generated', () => ({taskAttachmentsDownload: getBlobUrl}))
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
vi.mock(import('@/helpers/auth'), async importOriginal => ({
	...await importOriginal(),
	getAuthSessionEpoch: () => requestContext.sessionEpoch,
	getToken: () => null,
	getTokenIdentity: () => requestContext.identity,
}))
vi.mock(import('@/helpers/fetcher'), async importOriginal => ({
	...await importOriginal(),
	getApiV2BaseUrl: () => requestContext.apiV2BaseUrl,
}))

const attachment = {task_id: 5, id: 9}

beforeEach(() => {
	vi.useRealTimers()
	vi.unstubAllGlobals()
	requestContext.identity = {id: 1, type: 1}
	requestContext.sessionEpoch = 1
	requestContext.apiV2BaseUrl = 'https://identity-a.example/api/v2/'
	URL.createObjectURL = vi.fn(blob => (blob as Blob & {testUrl?: string}).testUrl ?? 'blob:real-attachment')
	queryClient.removeQueries({queryKey: attachmentKeys.blobs})
	getBlobUrl.mockReset()
	window.URL.revokeObjectURL = vi.fn()
})

describe('fetchAttachmentUrl', () => {
	it('returns an inert data url for svg, which would otherwise script in our origin', async () => {
		getBlobUrl.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml'})})

		expect(await fetchAttachmentUrl(attachment)).toMatch(/^data:image\/svg\+xml/)
		expect(URL.createObjectURL).not.toHaveBeenCalled()
	})

	it('returns an inert data url for svg with a charset parameter', async () => {
		getBlobUrl.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml; charset=utf-8'})})

		expect(await fetchAttachmentUrl(attachment)).toMatch(/^data:image\/svg\+xml/)
		expect(URL.createObjectURL).not.toHaveBeenCalled()
	})

	it('returns a blob url for html, which never reaches a rendering context', async () => {
		const html = Object.assign(new Blob(['<script />'], {type: 'text/html'}), {testUrl: 'blob:html'})
		getBlobUrl.mockResolvedValue({data: html})

		expect(await fetchAttachmentUrl(attachment)).toBe('blob:html')
		expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
	})

	it('rejects when reading fails instead of handing out a scriptable blob url', async () => {
		getBlobUrl.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml'})})
		vi.stubGlobal('FileReader', class {
			onerror: (() => void) | null = null
			error = new Error('read failed')
			readAsDataURL() {
				this.onerror?.()
			}
		})

		await expect(fetchAttachmentUrl(attachment)).rejects.toThrow('read failed')
		expect(URL.createObjectURL).not.toHaveBeenCalled()
	})

	it('rejects when the identity changes while the file is being read', async () => {
		getBlobUrl.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml'})})
		let startRead: (fire: () => void) => void = () => {}
		const reading = new Promise<() => void>(resolve => {
			startRead = resolve
		})
		vi.stubGlobal('FileReader', class {
			onload: (() => void) | null = null
			result = 'data:image/svg+xml;base64,PHN2ZyAvPg=='
			readAsDataURL() {
				startRead(() => this.onload?.())
			}
		})

		const pending = fetchAttachmentUrl(attachment)
		const finishRead = await reading
		requestContext.identity = {id: 2, type: 1}
		finishRead()

		await expect(pending).rejects.toMatchObject({
			name: 'AbortError',
			message: 'Client request context changed',
		})
		expect(URL.createObjectURL).not.toHaveBeenCalled()
	})

	it('rejects when the identity changed while the file was downloading', async () => {
		getBlobUrl.mockResolvedValue({data: new Blob(['%PDF'], {type: 'application/pdf'})})
		const pending = fetchAttachmentUrl(attachment)
		requestContext.identity = {id: 2, type: 1}

		await expect(pending).rejects.toMatchObject({
			name: 'AbortError',
			message: 'Client request context changed',
		})
	})

	it('returns a blob url for every other mime', async () => {
		const pdf = Object.assign(new Blob(['%PDF'], {type: 'application/pdf'}), {testUrl: 'blob:pdf'})
		getBlobUrl.mockResolvedValue({data: pdf})

		expect(await fetchAttachmentUrl(attachment)).toBe('blob:pdf')
		expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
	})

	it('hands every caller its own url over a single request', async () => {
		getBlobUrl.mockResolvedValue({data: Object.assign(new Blob(['bytes']), {testUrl: 'blob:a'})})

		expect(await fetchAttachmentUrl(attachment)).toBe('blob:a')
		expect(await fetchAttachmentUrl(attachment)).toBe('blob:a')

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
		expect(URL.createObjectURL).toHaveBeenCalledTimes(2)
	})
})

describe('fetchAttachmentBlob', () => {
	it('fetches once for repeated calls with the same key', async () => {
		const blob = new Blob(['bytes'])
		getBlobUrl.mockResolvedValue({data: blob})

		expect(await fetchAttachmentBlob(attachment)).toBe(blob)
		expect(await fetchAttachmentBlob(attachment)).toBe(blob)

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('shares one request between concurrent callers', async () => {
		const blob = new Blob(['bytes'])
		let resolveBlob: (result: {data: Blob}) => void = () => {}
		getBlobUrl.mockReturnValue(new Promise<{data: Blob}>(resolve => {
			resolveBlob = resolve
		}))

		const both = Promise.all([fetchAttachmentBlob(attachment), fetchAttachmentBlob(attachment)])
		resolveBlob({data: blob})

		expect(await both).toEqual([blob, blob])
		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('retries after a rejected fetch', async () => {
		const failed = new Error('nope')
		const blob = new Blob(['bytes'])
		getBlobUrl.mockRejectedValueOnce(failed).mockResolvedValueOnce({data: blob})

		await expect(fetchAttachmentBlob(attachment)).rejects.toThrow(failed)
		expect(await fetchAttachmentBlob(attachment)).toBe(blob)

		expect(getBlobUrl).toHaveBeenCalledTimes(2)
	})

	it('caches every preview size separately', async () => {
		getBlobUrl.mockResolvedValueOnce({data: new Blob(['original'])})
			.mockResolvedValueOnce({data: new Blob(['md'])})
			.mockResolvedValueOnce({data: new Blob(['lg'])})

		expect(await (await fetchAttachmentBlob(attachment)).text()).toBe('original')
		expect(await (await fetchAttachmentBlob(attachment, 'md')).text()).toBe('md')
		expect(await (await fetchAttachmentBlob(attachment, 'lg')).text()).toBe('lg')
		expect(await (await fetchAttachmentBlob(attachment, 'md')).text()).toBe('md')

		expect(getBlobUrl).toHaveBeenCalledTimes(3)
	})

	it('caches every attachment separately', async () => {
		getBlobUrl.mockResolvedValueOnce({data: new Blob(['nine'])})
			.mockResolvedValueOnce({data: new Blob(['ten'])})

		expect(await (await fetchAttachmentBlob({task_id: 5, id: 9})).text()).toBe('nine')
		expect(await (await fetchAttachmentBlob({task_id: 5, id: 10})).text()).toBe('ten')

		expect(getBlobUrl).toHaveBeenCalledTimes(2)
	})

	it('never revokes a url its caller may still show when the bytes are dropped', async () => {
		getBlobUrl.mockResolvedValue({data: Object.assign(new Blob(['bytes']), {testUrl: 'blob:a'})})
		const url = await fetchAttachmentUrl(attachment)

		queryClient.removeQueries({queryKey: attachmentKeys.blobs})

		expect(url).toBe('blob:a')
		expect(window.URL.revokeObjectURL).not.toHaveBeenCalled()
	})
})

describe('releaseAttachmentUrl', () => {
	it('revokes an object url', () => {
		releaseAttachmentUrl('blob:a')

		expect(window.URL.revokeObjectURL).toHaveBeenCalledExactlyOnceWith('blob:a')
	})

	it('leaves the inert svg data url alone, there is nothing to revoke', () => {
		releaseAttachmentUrl('data:image/svg+xml;base64,PHN2ZyAvPg==')
		releaseAttachmentUrl(undefined)

		expect(window.URL.revokeObjectURL).not.toHaveBeenCalled()
	})
})
