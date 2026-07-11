import {describe, it, expect} from 'vitest'

import {canPreviewImage, canPreviewPdf, canPreview} from './attachment'
import type {IAttachment} from '@/modelTypes/IAttachment'

function attachment(name: string, mime: string): IAttachment {
	return {file: {name, mime}} as unknown as IAttachment
}

describe('canPreviewPdf', () => {
	it('previews a real pdf', () => {
		expect(canPreviewPdf(attachment('doc.pdf', 'application/pdf'))).toBe(true)
	})

	it('refuses html bytes disguised as a pdf', () => {
		expect(canPreviewPdf(attachment('evil.pdf', 'text/html'))).toBe(false)
	})

	it('matches the mime case-insensitively', () => {
		expect(canPreviewPdf(attachment('doc.pdf', 'APPLICATION/PDF'))).toBe(true)
	})

	it('refuses a pdf mime without a .pdf name', () => {
		expect(canPreviewPdf(attachment('doc.txt', 'application/pdf'))).toBe(false)
	})
})

describe('canPreviewImage', () => {
	it('previews a real png', () => {
		expect(canPreviewImage(attachment('pic.png', 'image/png'))).toBe(true)
	})

	it('refuses html bytes disguised as a png', () => {
		expect(canPreviewImage(attachment('evil.png', 'text/html'))).toBe(false)
	})

	it('refuses svg since it can carry script', () => {
		expect(canPreviewImage(attachment('evil.jpg', 'image/svg+xml'))).toBe(false)
	})
})

describe('canPreview', () => {
	it('refuses html bytes disguised as a pdf', () => {
		expect(canPreview(attachment('evil.pdf', 'text/html'))).toBe(false)
	})
})
