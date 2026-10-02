import {beforeEach, describe, expect, it, vi} from 'vitest'
import type {MutationOptions} from '@tanstack/vue-query'

import {queryClient} from '@/client/queryClient'

const http = vi.hoisted(() => ({
	get: vi.fn(),
	post: vi.fn(),
	put: vi.fn(),
	patch: vi.fn(),
	delete: vi.fn(),
}))

vi.mock('@/client/generated/client.gen', () => ({client: http}))

import {
	createManagedUserMutation,
	deleteManagedUserMutation,
	importStatusQuery,
	manageKeys,
	manageUserProjectsQuery,
	manageUsersQuery,
	previewUserImportMutation,
	runUserImportMutation,
	setManagedUserPasswordMutation,
	transferProjectsMutation,
	updateManagedUserMutation,
	uploadUserListMutation,
} from './manage'
import type {ImportStatus, ManagedUser} from './manage'

const user: ManagedUser = {
	id: 2,
	username: 'user2',
	name: 'User Two',
	email: 'user2@example.com',
	language: 'en',
	job_title: 'Engineer',
	department: 'IT',
	status: 0,
	is_admin: false,
	source: 'local',
	issuer: 'local',
	must_change_password: false,
	profile_editable: true,
	created: '2026-01-01T00:00:00Z',
}

// Through the real mutation cache, like labels.test.ts, so onSettled runs too.
function runMutation<TData, TVars, TContext>(
	options: MutationOptions<TData, Error, TVars, TContext>,
	vars: TVars,
): Promise<TData> {
	return queryClient.getMutationCache().build(queryClient, options).execute(vars)
}

describe('manage queries', () => {
	beforeEach(() => {
		queryClient.clear()
		for (const fn of Object.values(http)) {
			fn.mockReset()
		}
	})

	it('requests a page of users with the filters and the page size', async () => {
		http.get.mockResolvedValue({data: {items: [user], total: 1, page: 2, per_page: 50, total_pages: 1}})

		const page = await queryClient.fetchQuery(manageUsersQuery({page: 2, q: 'jane', status: 2, source: 'entra'}))

		expect(page.items).toEqual([user])
		expect(http.get).toHaveBeenCalledWith(expect.objectContaining({
			url: '/manage/users',
			query: {page: 2, per_page: 50, q: 'jane', status: 2, source: 'entra'},
			throwOnError: true,
		}))
	})

	it('leaves unset filters out of the request, status 0 included only when chosen', async () => {
		http.get.mockResolvedValue({data: {items: [], total: 0, page: 1, per_page: 50, total_pages: 0}})

		await queryClient.fetchQuery(manageUsersQuery({page: 1}))
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({query: {page: 1, per_page: 50}}))

		await queryClient.fetchQuery(manageUsersQuery({page: 1, status: 0}))
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({query: {page: 1, per_page: 50, status: 0}}))
	})

	it('keys lists by filter so a filter change is a different cache entry', () => {
		expect(manageKeys.userList({page: 1})).not.toEqual(manageKeys.userList({page: 2}))
		expect(manageKeys.userList({page: 1, q: 'a'})).not.toEqual(manageKeys.userList({page: 1}))
		expect(manageKeys.userList({page: 1})).toEqual(expect.arrayContaining(manageKeys.users()))
	})

	it('loads the projects of a user', async () => {
		http.get.mockResolvedValue({data: {items: [{id: 5, title: 'P', is_archived: false, parent_project_id: 0}]}})

		const projects = await queryClient.fetchQuery(manageUserProjectsQuery(3))

		expect(projects).toHaveLength(1)
		expect(http.get).toHaveBeenCalledWith(expect.objectContaining({url: '/manage/users/3/projects'}))
	})

	it('polls the import status only while a run is going', () => {
		const query = importStatusQuery()
		const refetch = query.refetchInterval as (q: {state: {data?: Partial<ImportStatus>}}) => number | false

		expect(refetch({state: {data: {running: true}}})).toBe(2000)
		expect(refetch({state: {data: {running: false}}})).toBe(false)
		expect(refetch({state: {}})).toBe(false)
		expect((importStatusQuery(false).refetchInterval as typeof refetch)({state: {data: {running: true}}})).toBe(false)
	})

	describe('mutations', () => {
		it('updates a profile with a JSON body and no id in it', async () => {
			http.put.mockResolvedValue({data: user})

			await runMutation(updateManagedUserMutation(), {id: 2, name: 'New', email: 'new@example.com', language: 'de'})

			expect(http.put).toHaveBeenCalledWith(expect.objectContaining({
				url: '/manage/users/2',
				body: {name: 'New', email: 'new@example.com', language: 'de'},
			}))
		})

		it('invalidates every user list once a mutation settled, and nothing else', async () => {
			http.put.mockResolvedValue({data: user})
			const spy = vi.spyOn(queryClient, 'invalidateQueries')

			await runMutation(updateManagedUserMutation(), {id: 2, name: 'n', email: 'e@example.com'})

			expect(spy).toHaveBeenCalledExactlyOnceWith({queryKey: manageKeys.users()})
			spy.mockRestore()
		})
		it('creates a user', async () => {
			http.post.mockResolvedValue({data: user})

			await runMutation(createManagedUserMutation(), {username: 'u', email: 'u@example.com', password: 'secret-pass', require_change: true})

			expect(http.post).toHaveBeenCalledWith(expect.objectContaining({
				url: '/manage/users',
				body: {username: 'u', email: 'u@example.com', password: 'secret-pass', require_change: true},
			}))
		})

		it('sets a password and says whether a change is required', async () => {
			http.patch.mockResolvedValue({data: user})

			await runMutation(setManagedUserPasswordMutation(), {id: 2, new_password: 'secret-pass', require_change: false})

			expect(http.patch).toHaveBeenCalledWith(expect.objectContaining({
				url: '/manage/users/2/password',
				body: {new_password: 'secret-pass', require_change: false},
			}))
		})

		it('deletes with the chosen mode in the query', async () => {
			http.delete.mockResolvedValue({data: undefined})

			await runMutation(deleteManagedUserMutation(), {id: 4, mode: 'now'})

			expect(http.delete).toHaveBeenCalledWith(expect.objectContaining({url: '/manage/users/4', query: {mode: 'now'}}))
		})

		it('transfers selected projects, or all of them when none are named', async () => {
			http.post.mockResolvedValue({data: {transferred: 2}})

			await runMutation(transferProjectsMutation(), {id: 3, new_owner_id: 2, project_ids: [7, 8]})
			expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({
				url: '/manage/users/3/transfer-projects',
				body: {new_owner_id: 2, project_ids: [7, 8]},
			}))

			await runMutation(transferProjectsMutation(), {id: 3, new_owner_id: 2})
			expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({body: {new_owner_id: 2}}))
		})

		it('uploads the list as multipart under the field the server reads', async () => {
			http.put.mockResolvedValue({data: {rows: 3, skipped_invalid: 0, skipped_duplicate: 0}})
			const file = new File(['id,mail,displayName\n1,a@example.com,A\n'], 'user_list.csv', {type: 'text/csv'})

			await runMutation(uploadUserListMutation(), file)

			const call = http.put.mock.calls[0][0]
			expect(call.url).toBe('/manage/user-import/file')
			expect(call.body).toBeInstanceOf(FormData)
			expect((call.body as FormData).get('list')).toBe(file)
			// The browser has to set the multipart boundary itself.
			expect(call.headers['Content-Type']).toBeNull()
		})

		it('previews and runs the import', async () => {
			http.post.mockResolvedValue({data: {counts: {}}})

			await runMutation(previewUserImportMutation(), undefined)
			expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({url: '/manage/user-import/preview'}))

			await runMutation(runUserImportMutation(), undefined)
			expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({url: '/manage/user-import/run'}))
		})

		it('refreshes the import status after a run is started', async () => {
			http.post.mockResolvedValue({data: {message: 'ok'}})
			const spy = vi.spyOn(queryClient, 'invalidateQueries')

			await runMutation(runUserImportMutation(), undefined)

			expect(spy).toHaveBeenCalledWith({queryKey: manageKeys.importStatus()})
			spy.mockRestore()
		})
		it('surfaces a server error to the caller instead of swallowing it', async () => {
			http.put.mockRejectedValue({status: 403, code: 1042, detail: 'The profile of this user is managed externally.'})

			await expect(runMutation(updateManagedUserMutation(), {id: 14, name: 'x', email: 'x@example.com'})).rejects.toMatchObject({code: 1042})
		})
	})
})
