import {describe, it, expect, vi, afterEach} from 'vitest'

import AttachmentService from './attachment'

function serviceWithBlobResponse(blob: Blob) {
	const service = new AttachmentService()
	service.http = vi.fn().mockResolvedValue({data: blob}) as unknown as typeof service.http
	return service
}

describe('getBlobUrl', () => {
	afterEach(() => {
		vi.restoreAllMocks()
	})

	it('keeps the mime type of the fetched blob', async () => {
		// A blob url without a type downloads instead of rendering when used as iframe src
		const service = serviceWithBlobResponse(new Blob(['%PDF-1.4'], {type: 'application/pdf'}))
		const createObjectURL = vi.spyOn(window.URL, 'createObjectURL').mockReturnValue('blob:mock')

		const url = await service.getBlobUrl('/tasks/1/attachments/1')

		expect(url).toBe('blob:mock')
		const blob = createObjectURL.mock.calls[0][0] as Blob
		expect(blob.type).toBe('application/pdf')
		expect(blob.size).toBeGreaterThan(0)
	})

	it('converts svg blobs to data urls', async () => {
		const service = serviceWithBlobResponse(new Blob(['<svg xmlns="http://www.w3.org/2000/svg"/>'], {type: 'image/svg+xml'}))

		const url = await service.getBlobUrl('/tasks/1/attachments/2')

		expect(url).toMatch(/^data:image\/svg\+xml/)
	})
})
