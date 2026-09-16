import {describe, it, expect, vi, beforeEach} from 'vitest'

import EmailUpdateService from './emailUpdate'
import EmailUpdateModel from '@/models/emailUpdate'

const {put, del, post} = vi.hoisted(() => ({
	put: vi.fn(),
	del: vi.fn(),
	post: vi.fn(),
}))

const http = {
	put,
	delete: del,
	post,
	interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}},
}

vi.mock('@/helpers/fetcher', async (importOriginal) => ({
	...await importOriginal<typeof import('@/helpers/fetcher')>(),
	HTTPFactory: () => http,
	AuthenticatedHTTPFactory: () => http,
}))

describe('EmailUpdateService', () => {
	beforeEach(() => {
		// the real apiV2Url derives the v2 base from this
		window.API_URL = 'http://localhost/api/v1/'
		put.mockReset()
		del.mockReset()
		post.mockReset()
	})

	it('puts the new email to the v2 endpoint', async () => {
		const model = new EmailUpdateModel({
			newEmail: 'new@example.com',
			password: 'secret',
		})

		await new EmailUpdateService().update(model)

		expect(put).toHaveBeenCalledOnce()
		expect(put.mock.calls[0][0]).toMatch(/\/api\/v2\/user\/settings\/email$/)

		// snake-casing is the shared request interceptor's job, so the payload goes out camelCased
		const payload = put.mock.calls[0][1]
		expect(payload).toEqual({newEmail: 'new@example.com', password: 'secret'})
		// AbstractModel fields like maxPermission make v2 reject the body with 422
		expect(payload).not.toHaveProperty('maxPermission')
		expect(payload).not.toHaveProperty('max_permission')
	})

	it('deletes the pending change on the v2 endpoint', async () => {
		await new EmailUpdateService().cancel()

		expect(del).toHaveBeenCalledOnce()
		expect(del.mock.calls[0][0]).toMatch(/\/api\/v2\/user\/settings\/email$/)
	})

	it('posts to the v2 resend endpoint', async () => {
		await new EmailUpdateService().resend()

		expect(post).toHaveBeenCalledOnce()
		expect(post.mock.calls[0][0]).toMatch(/\/api\/v2\/user\/settings\/email\/resend$/)
	})
})
