import {describe, expect, it} from 'vitest'

import {expectBlob} from './blobResponse'

describe('expectBlob', () => {
	it('returns the blob unchanged', () => {
		const blob = new Blob(['bytes'])

		expect(expectBlob(blob, 'Attachment')).toBe(blob)
	})

	it('names the resource when the response is not a blob', () => {
		expect(() => expectBlob({detail: 'not found'}, 'Background'))
			.toThrowError('Background response was not a blob')
	})
})
