import {describe, it, expect} from 'vitest'

import {canPreviewAudio, canPreviewImage, canPreviewPdf, canPreviewVideo} from './attachment'
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

describe('canPreviewAudio', () => {
	it('previews an mp3', () => {
		expect(canPreviewAudio(attachment('memo.mp3', 'audio/mpeg'))).toBe(true)
	})

	it('previews an ogg regardless of the file name', () => {
		expect(canPreviewAudio(attachment('19-40-43', 'audio/ogg'))).toBe(true)
	})

	it('matches the mime case-insensitively', () => {
		expect(canPreviewAudio(attachment('memo.mp3', 'AUDIO/MPEG'))).toBe(true)
	})

	it('refuses a non-audio mime', () => {
		expect(canPreviewAudio(attachment('memo.mp3', 'text/html'))).toBe(false)
	})
})

describe('canPreviewVideo', () => {
	it('previews a real mp4', () => {
		expect(canPreviewVideo(attachment('clip.mp4', 'video/mp4'))).toBe(true)
	})

	it('previews a real webm', () => {
		expect(canPreviewVideo(attachment('clip.webm', 'video/webm'))).toBe(true)
	})

	it('refuses text bytes disguised as an mp4', () => {
		expect(canPreviewVideo(attachment('evil.mp4', 'text/plain'))).toBe(false)
	})

	it('refuses a video mime without a video suffix', () => {
		expect(canPreviewVideo(attachment('clip.txt', 'video/mp4'))).toBe(false)
	})

	it('matches the mime case-insensitively', () => {
		expect(canPreviewVideo(attachment('clip.MP4', 'VIDEO/MP4'))).toBe(true)
	})

	it('refuses audio ogg', () => {
		expect(canPreviewVideo(attachment('song.ogg', 'audio/ogg'))).toBe(false)
	})

	it('previews video ogg', () => {
		expect(canPreviewVideo(attachment('clip.ogg', 'video/ogg'))).toBe(true)
	})

	it('refuses a video mime with an unsupported .mkv suffix', () => {
		expect(canPreviewVideo(attachment('clip.mkv', 'video/x-matroska'))).toBe(false)
	})
})
