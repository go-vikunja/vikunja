import {describe, it, expect, beforeEach, vi} from 'vitest'

import {clearAttachmentBlobCache, fetchAttachmentBlobUrl, uploadFilesForEditor} from './attachments'
import {PREVIEW_SIZE} from '@/services/attachment'

const {getBlobUrl} = vi.hoisted(() => ({getBlobUrl: vi.fn()}))

vi.mock('@/services/attachment', async importOriginal => ({
	...await importOriginal<typeof import('@/services/attachment')>(),
	default: class {
		getBlobUrl = getBlobUrl
	},
}))

const attachment = {taskId: 5, id: 9}

beforeEach(() => {
	clearAttachmentBlobCache()
	getBlobUrl.mockReset()
	window.URL.revokeObjectURL = vi.fn()
})

describe('fetchAttachmentBlobUrl', () => {
	it('fetches once for repeated calls with the same key', async () => {
		getBlobUrl.mockResolvedValue('blob:a')

		expect(await fetchAttachmentBlobUrl(attachment)).toBe('blob:a')
		expect(await fetchAttachmentBlobUrl(attachment)).toBe('blob:a')

		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('shares one request between concurrent callers', async () => {
		let resolveBlobUrl: (url: string) => void = () => {}
		getBlobUrl.mockReturnValue(new Promise<string>(resolve => {
			resolveBlobUrl = resolve
		}))

		const both = Promise.all([fetchAttachmentBlobUrl(attachment), fetchAttachmentBlobUrl(attachment)])
		resolveBlobUrl('blob:a')

		expect(await both).toEqual(['blob:a', 'blob:a'])
		expect(getBlobUrl).toHaveBeenCalledTimes(1)
	})

	it('retries after a rejected fetch', async () => {
		const failed = new Error('nope')
		getBlobUrl.mockRejectedValueOnce(failed).mockResolvedValueOnce('blob:a')

		await expect(fetchAttachmentBlobUrl(attachment)).rejects.toThrow(failed)
		expect(await fetchAttachmentBlobUrl(attachment)).toBe('blob:a')

		expect(getBlobUrl).toHaveBeenCalledTimes(2)
	})

	it('caches every preview size separately', async () => {
		getBlobUrl.mockResolvedValueOnce('blob:original')
			.mockResolvedValueOnce('blob:md')
			.mockResolvedValueOnce('blob:lg')

		expect(await fetchAttachmentBlobUrl(attachment)).toBe('blob:original')
		expect(await fetchAttachmentBlobUrl(attachment, PREVIEW_SIZE.MD)).toBe('blob:md')
		expect(await fetchAttachmentBlobUrl(attachment, PREVIEW_SIZE.LG)).toBe('blob:lg')
		expect(await fetchAttachmentBlobUrl(attachment, PREVIEW_SIZE.MD)).toBe('blob:md')

		expect(getBlobUrl).toHaveBeenCalledTimes(3)
	})

	it('caches every attachment separately', async () => {
		getBlobUrl.mockResolvedValueOnce('blob:a').mockResolvedValueOnce('blob:b')

		expect(await fetchAttachmentBlobUrl({taskId: 5, id: 9})).toBe('blob:a')
		expect(await fetchAttachmentBlobUrl({taskId: 5, id: 10})).toBe('blob:b')

		expect(getBlobUrl).toHaveBeenCalledTimes(2)
	})
})

describe('clearAttachmentBlobCache', () => {
	it('revokes the cached urls and refetches afterwards', async () => {
		getBlobUrl.mockResolvedValueOnce('blob:a').mockResolvedValueOnce('blob:b')
		await fetchAttachmentBlobUrl(attachment)

		clearAttachmentBlobCache()

		expect(window.URL.revokeObjectURL).toHaveBeenCalledWith('blob:a')
		expect(await fetchAttachmentBlobUrl(attachment)).toBe('blob:b')
	})
})

describe('uploadFilesForEditor', () => {
	it('collects the url of every uploaded file', async () => {
		const upload = vi.fn((file: File, onSuccess: (url: string) => void) => {
			onSuccess(`https://vikunja.io/${file.name}`)
			return Promise.resolve()
		})

		const urls = await uploadFilesForEditor(upload, [new File([''], 'a.png'), new File([''], 'b.png')])

		expect(urls).toEqual(['https://vikunja.io/a.png', 'https://vikunja.io/b.png'])
	})

	it('rejects when an upload fails instead of leaving the promise dangling', async () => {
		const failed = new Error('failed to save file: no space left on device')

		await expect(uploadFilesForEditor(() => Promise.reject(failed), [new File([''], 'a.png')]))
			.rejects.toThrow(failed)
	})

	it('handles a FileList, which has no forEach', async () => {
		const file = new File([''], 'a.png')
		const files = {0: file, length: 1, [Symbol.iterator]: [file][Symbol.iterator].bind([file])} as unknown as FileList

		const urls = await uploadFilesForEditor((f, onSuccess) => {
			onSuccess(`https://vikunja.io/${f.name}`)
			return Promise.resolve()
		}, files)

		expect(urls).toEqual(['https://vikunja.io/a.png'])
	})
})
