import {describe, it, expect, vi, beforeEach} from 'vitest'

const get = vi.fn()
const patch = vi.fn()
vi.mock('@/helpers/fetcher', () => ({
	AuthenticatedHTTPFactory: () => ({get, patch}),
}))

import {listAdminProjects, reassignProjectOwner} from './projectService'

describe('admin projectService', () => {
	beforeEach(() => {
		get.mockReset()
		patch.mockReset()
	})

	it('GETs /admin/projects with pagination params', async () => {
		get.mockResolvedValue({data: []})
		await listAdminProjects({page: 2, perPage: 25})
		expect(get).toHaveBeenCalledWith('/admin/projects', {
			params: {page: 2, per_page: 25},
		})
	})

	it('camel-cases project rows', async () => {
		get.mockResolvedValue({data: [{id: 1, title: 'p1', owner_id: 3}]})
		const projects = await listAdminProjects()
		expect(projects[0]).toMatchObject({id: 1, title: 'p1', ownerId: 3})
	})

	it('PATCHes owner reassignment with snake_case body', async () => {
		patch.mockResolvedValue({data: {id: 1, owner_id: 9}})
		await reassignProjectOwner(1, 9)
		expect(patch).toHaveBeenCalledWith('/admin/projects/1/owner', {owner_id: 9})
	})
})
