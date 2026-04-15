import {describe, it, expect, vi, beforeEach} from 'vitest'

const get = vi.fn()
const patch = vi.fn()
vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => ({get, patch}),
}))

import {listAdminUsers, setAdmin} from './userService'

describe('admin userService', () => {
	beforeEach(() => {
		get.mockReset()
		patch.mockReset()
	})

	it('GETs /admin/users with snake-cased params', async () => {
		get.mockResolvedValue({data: []})
		await listAdminUsers({s: 'bob', page: 2, perPage: 25})
		expect(get).toHaveBeenCalledWith('/admin/users', {
			params: {s: 'bob', page: 2, per_page: 25},
		})
	})

	it('camel-cases each user row', async () => {
		get.mockResolvedValue({data: [
			{id: 1, username: 'u1', is_admin: true, status: 0},
		]})
		const users = await listAdminUsers()
		expect(users[0].isAdmin).toBe(true)
		expect(users[0].status).toBe(0)
	})

	it('PATCHes /admin/users/:id/admin with is_admin flag', async () => {
		patch.mockResolvedValue({data: {id: 7, is_admin: true, status: 0}})
		const out = await setAdmin(7, true)
		expect(patch).toHaveBeenCalledWith('/admin/users/7/admin', {is_admin: true})
		expect(out.isAdmin).toBe(true)
	})
})
