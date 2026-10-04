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
	canWriteRisk,
	changeRiskStatusMutation,
	createRiskMutation,
	deleteRiskMutation,
	fetchAllRisks,
	isRiskStatusConflict,
	MAX_RISK_EXPORT,
	riskHistoryQuery,
	riskKeys,
	riskListQuery,
	riskQuery,
	risksQuery,
	updateRiskMutation,
} from './risks'
import type {Risk, RiskInput} from './risks'

function risk(id: number, overrides: Partial<Risk> = {}): Risk {
	return {
		id, project_id: 1, title: `Risk ${id}`, description: '', category: '', probability: 3, impact: 3, score: 9, rating: 'medium',
		owner_id: 0, owner: null, mitigation: '', contingency: '', status: 'open', identified_date: null, due_date: null,
		closed_at: null, closed_by: null, resolution: '', created_by: null, created: '2026-01-01T00:00:00Z', updated: '2026-01-01T00:00:00Z',
		...overrides,
	}
}

const input: RiskInput = {
	title: 'Supplier delay', description: '', category: 'Schedule', probability: 4, impact: 5, owner_id: 0,
	mitigation: '', contingency: '', identified_date: null, due_date: null,
}

function runMutation<TData, TVars, TContext>(options: MutationOptions<TData, Error, TVars, TContext>, vars: TVars): Promise<TData> {
	return queryClient.getMutationCache().build(queryClient, options).execute(vars)
}

describe('riskListQuery', () => {
	const base = {page: 2, perPage: 50}

	it('sends paging and nothing else for no filters', () => {
		expect(riskListQuery(base, false)).toEqual({page: 2, per_page: 50})
	})

	it('sends every filter once, repeatable ones as arrays', () => {
		expect(riskListQuery({
			...base, q: 'delay', projectIds: [3, 4], statuses: ['open', 'closed'], ratings: ['high'], ownerIds: [0, 9],
			categories: ['Schedule'], overdue: true, includeArchived: true, sortBy: 'score', orderBy: 'desc',
		}, false)).toEqual({
			page: 2, per_page: 50, q: 'delay', project_id: [3, 4], status: ['open', 'closed'], rating: ['high'], owner_id: [0, 9],
			category: ['Schedule'], overdue: true, include_archived: true, sort_by: ['score'], order_by: ['desc'],
		})
	})

	it('a project page has its project in the path, so the project filters are not sent', () => {
		const query = riskListQuery({...base, projectIds: [3], includeArchived: true, statuses: ['open']}, true)
		expect(query).toEqual({page: 2, per_page: 50, status: ['open']})
	})

	it('a sort without an order is ascending', () => {
		expect(riskListQuery({...base, sortBy: 'title'}, false)).toMatchObject({sort_by: ['title'], order_by: ['asc']})
	})

	it('leaves out empty lists and false flags', () => {
		expect(riskListQuery({...base, q: '', statuses: [], ratings: [], overdue: false, includeArchived: false}, false)).toEqual({page: 2, per_page: 50})
	})
})

describe('risk queries', () => {
	beforeEach(() => {
		queryClient.clear()
		for (const fn of Object.values(http)) fn.mockReset()
	})

	it('lists all projects, or one project, with its own cache key', async () => {
		http.get.mockResolvedValue({data: {items: [risk(1)], total: 1, page: 1, per_page: 50, total_pages: 1}})
		const params = {page: 1, perPage: 50, statuses: ['open' as const]}

		await queryClient.fetchQuery(risksQuery(null, params))
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({url: '/risks', query: {page: 1, per_page: 50, status: ['open']}, throwOnError: true}))

		await queryClient.fetchQuery(risksQuery(7, params))
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({url: '/projects/7/risks'}))

		expect(riskKeys.list(null, params)).not.toEqual(riskKeys.list(7, params))
		expect(riskKeys.list(7, params)).not.toEqual(riskKeys.list(7, {...params, page: 2}))
		expect(riskKeys.list(7, params)).toEqual(expect.arrayContaining(riskKeys.lists()))
	})

	it('reads one risk and its history', async () => {
		http.get.mockResolvedValueOnce({data: risk(5)})
		expect((await queryClient.fetchQuery(riskQuery(5))).id).toBe(5)
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({url: '/risks/5'}))

		http.get.mockResolvedValueOnce({data: {items: [{id: 1, risk_id: 5, from_status: '', to_status: 'open', note: '', changed_by: null, created: ''}], total: 1}})
		const history = await queryClient.fetchQuery(riskHistoryQuery(5))
		expect(history).toHaveLength(1)
		expect(http.get).toHaveBeenLastCalledWith(expect.objectContaining({url: '/risks/5/history'}))
	})
})

describe('fetchAllRisks', () => {
	beforeEach(() => {
		for (const fn of Object.values(http)) fn.mockReset()
	})

	it('follows the pages until it has everything', async () => {
		http.get
			.mockResolvedValueOnce({data: {items: [risk(1), risk(2)], total: 3, page: 1, per_page: 2, total_pages: 2}})
			.mockResolvedValueOnce({data: {items: [risk(3)], total: 3, page: 2, per_page: 2, total_pages: 2}})

		const result = await fetchAllRisks(null, {statuses: ['open']})

		expect(result.risks.map(r => r.id)).toEqual([1, 2, 3])
		expect(result.truncated).toBe(false)
		expect(http.get).toHaveBeenCalledTimes(2)
		expect(http.get.mock.calls[1][0].query.page).toBe(2)
	})

	it('a project is read through its own path', async () => {
		http.get.mockResolvedValue({data: {items: [risk(1)], total: 1, page: 1, per_page: 1000, total_pages: 1}})
		await fetchAllRisks(4, {})
		expect(http.get).toHaveBeenCalledWith(expect.objectContaining({url: '/projects/4/risks'}))
	})

	it('stops at the limit and says so', async () => {
		http.get.mockResolvedValue({data: {items: [risk(1), risk(2), risk(3)], total: 10, page: 1, per_page: 3, total_pages: 4}})

		const result = await fetchAllRisks(null, {}, 5)

		expect(result.risks).toHaveLength(5)
		expect(result.truncated).toBe(true)
		expect(result.total).toBe(10)
	})

	it('does not loop forever on a server that returns an empty page', async () => {
		http.get.mockResolvedValue({data: {items: [], total: 7, page: 1, per_page: 1000, total_pages: 1}})
		const result = await fetchAllRisks(null, {})
		expect(result.risks).toEqual([])
		expect(http.get).toHaveBeenCalledTimes(1)
	})

	it('the export limit is 5000', () => {
		expect(MAX_RISK_EXPORT).toBe(5000)
	})
})

describe('risk mutations', () => {
	beforeEach(() => {
		queryClient.clear()
		for (const fn of Object.values(http)) fn.mockReset()
	})

	it('creates a risk in the project of the path, without the project in the body', async () => {
		http.post.mockResolvedValue({data: risk(9)})

		await runMutation(createRiskMutation(), {...input, projectId: 3})

		expect(http.post).toHaveBeenCalledWith(expect.objectContaining({url: '/projects/3/risks', body: input}))
	})

	it('updates with the id in the path and not in the body', async () => {
		http.put.mockResolvedValue({data: risk(9)})

		await runMutation(updateRiskMutation(), {...input, id: 9})

		expect(http.put).toHaveBeenCalledWith(expect.objectContaining({url: '/risks/9', body: input}))
	})

	it('changes the status, with the note only when there is one', async () => {
		http.post.mockResolvedValue({data: risk(9, {status: 'closed'})})

		await runMutation(changeRiskStatusMutation(), {id: 9, status: 'closed', note: 'Delivered'})
		expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({url: '/risks/9/status', body: {status: 'closed', note: 'Delivered'}}))

		await runMutation(changeRiskStatusMutation(), {id: 9, status: 'open'})
		expect(http.post).toHaveBeenLastCalledWith(expect.objectContaining({body: {status: 'open'}}))
	})

	it('deletes', async () => {
		http.delete.mockResolvedValue({data: undefined})
		await runMutation(deleteRiskMutation(), 9)
		expect(http.delete).toHaveBeenCalledWith(expect.objectContaining({url: '/risks/9'}))
	})

	it('invalidates every risk list, detail and history once a mutation settled, and nothing else', async () => {
		http.post.mockResolvedValue({data: risk(9)})
		const spy = vi.spyOn(queryClient, 'invalidateQueries')

		await runMutation(changeRiskStatusMutation(), {id: 9, status: 'closed'})

		expect(spy).toHaveBeenCalledExactlyOnceWith({queryKey: riskKeys.all})
		spy.mockRestore()
	})

	it('hands a server error to the caller', async () => {
		http.post.mockRejectedValue({status: 409, code: 20004, detail: 'changed by somebody else'})
		await expect(runMutation(changeRiskStatusMutation(), {id: 9, status: 'closed'})).rejects.toMatchObject({code: 20004})
	})
})

describe('isRiskStatusConflict', () => {
	it('recognises the conflict by code, status or response', () => {
		expect(isRiskStatusConflict({code: 20004})).toBe(true)
		expect(isRiskStatusConflict({status: 409})).toBe(true)
		expect(isRiskStatusConflict({response: {status: 409}})).toBe(true)
		expect(isRiskStatusConflict({response: {data: {code: 20004}}})).toBe(true)
	})

	it('does not take other errors for it', () => {
		expect(isRiskStatusConflict({code: 20002, status: 400})).toBe(false)
		expect(isRiskStatusConflict(new Error('offline'))).toBe(false)
		expect(isRiskStatusConflict(null)).toBe(false)
		expect(isRiskStatusConflict('409')).toBe(false)
	})
})

describe('canWriteRisk', () => {
	it('needs write permission, and is cautious until the permission is known', () => {
		expect(canWriteRisk({max_permission: 0})).toBe(false)
		expect(canWriteRisk({max_permission: 1})).toBe(true)
		expect(canWriteRisk({max_permission: 2})).toBe(true)
		expect(canWriteRisk({})).toBe(false)
	})
})
