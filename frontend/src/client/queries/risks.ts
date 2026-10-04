import {mutationOptions, queryOptions} from '@tanstack/vue-query'
import type {QueryClient} from '@tanstack/vue-query'

import {client} from '@/client/generated/client.gen'
import {API_MAX_PER_PAGE} from '@/client/queries/pagination'
import type {Paginated} from '@/client/queries/pagination'
import type {RiskRating, RiskStatus} from '@/helpers/riskRating'

// /api/v2/risks and /api/v2/projects/{project_id}/risks are not in the generated SDK until
// `mage generate:frontend-client` was run (needs Go). Until then they go through the same generated
// runtime client, so authentication, token refresh and the identity fencing in client/http.ts apply.
// The wire types mirror pkg/models/risk.go and pkg/routes/api/v2/risks.go. After generating, replace the
// bodies of the functions below with the generated operations and delete the local types.

export interface RiskUser {
	id: number
	name: string
	username: string
}

export interface Risk {
	id: number
	project_id: number
	title: string
	description: string
	category: string
	probability: number
	impact: number
	score: number
	rating: RiskRating
	owner_id: number
	owner: RiskUser | null
	mitigation: string
	contingency: string
	status: RiskStatus
	identified_date: string | null
	due_date: string | null
	closed_at: string | null
	closed_by: RiskUser | null
	resolution: string
	created_by: RiskUser | null
	created: string
	updated: string
	// Only in the response of a single risk: 0 read, 1 read and write, 2 admin.
	max_permission?: number
}

// A risk can be changed with write access to its project.
export function canWriteRisk(risk: Pick<Risk, 'max_permission'>): boolean {
	return (risk.max_permission ?? -1) >= 1
}

export interface RiskHistoryEntry {
	id: number
	risk_id: number
	from_status: RiskStatus | ''
	to_status: RiskStatus
	note: string
	changed_by: RiskUser | null
	created: string
}

// What a list can be narrowed by. The filters live in the URL of the pages, so this is also the shape
// that is written to and read from the query string (see riskFilterQuery.ts).
export interface RiskFilters {
	q?: string
	// Only for the register over all projects; a project page has its own project.
	projectIds?: number[]
	statuses?: RiskStatus[]
	ratings?: RiskRating[]
	// 0 stands for "no owner"
	ownerIds?: number[]
	categories?: string[]
	overdue?: boolean
	includeArchived?: boolean
	sortBy?: RiskSortKey
	orderBy?: 'asc' | 'desc'
}

export type RiskSortKey = 'score' | 'probability' | 'impact' | 'due_date' | 'status' | 'title' | 'created' | 'updated' | 'id'

export interface RiskListParams extends RiskFilters {
	page: number
	perPage: number
}

export const RISK_PAGE_SIZE = 50

// What can be edited. The status is changed by its own operation.
export interface RiskInput {
	title: string
	description: string
	category: string
	probability: number
	impact: number
	owner_id: number
	mitigation: string
	contingency: string
	identified_date: string | null
	due_date: string | null
}

export const ERR_CODE_RISK_STATUS_CONFLICT = 20004

// Somebody else changed the status between the moment the risk was loaded and the change.
export function isRiskStatusConflict(e: unknown): boolean {
	if (!e || typeof e !== 'object') {
		return false
	}
	const candidate = e as {code?: unknown, status?: unknown, response?: {status?: unknown, data?: {code?: unknown}}}
	return candidate.code === ERR_CODE_RISK_STATUS_CONFLICT ||
		candidate.status === 409 ||
		candidate.response?.status === 409 ||
		candidate.response?.data?.code === ERR_CODE_RISK_STATUS_CONFLICT
}

export const riskKeys = {
	all: ['risks'] as const,
	lists: () => [...riskKeys.all, 'list'] as const,
	list: (scope: number | null, params: RiskListParams) => [...riskKeys.lists(), scope, params] as const,
	detail: (id: number) => [...riskKeys.all, 'detail', id] as const,
	history: (id: number) => [...riskKeys.all, 'history', id] as const,
}

async function request<T>(method: 'get' | 'post' | 'put' | 'patch' | 'delete', url: string, options: Record<string, unknown> = {}): Promise<T> {
	const result = await client[method]({url, ...options, throwOnError: true} as never) as {data: T}
	return result.data
}

// The query string of a list. Empty filters are left out, repeatable ones go out once per value.
export function riskListQuery(params: RiskListParams, projectScope: boolean): Record<string, unknown> {
	const query: Record<string, unknown> = {page: params.page, per_page: params.perPage}
	if (params.q) query.q = params.q
	if (!projectScope && params.projectIds?.length) query.project_id = params.projectIds
	if (params.statuses?.length) query.status = params.statuses
	if (params.ratings?.length) query.rating = params.ratings
	if (params.ownerIds?.length) query.owner_id = params.ownerIds
	if (params.categories?.length) query.category = params.categories
	if (params.overdue) query.overdue = true
	if (!projectScope && params.includeArchived) query.include_archived = true
	if (params.sortBy) {
		query.sort_by = [params.sortBy]
		query.order_by = [params.orderBy ?? 'asc']
	}
	return query
}

function listUrl(projectId: number | null): string {
	return projectId === null ? '/risks' : `/projects/${projectId}/risks`
}

// projectId null is the register over all projects.
export function risksQuery(projectId: number | null, params: RiskListParams) {
	return queryOptions({
		queryKey: riskKeys.list(projectId, params),
		queryFn: () => request<Paginated<Risk>>('get', listUrl(projectId), {query: riskListQuery(params, projectId !== null)}),
		staleTime: 30 * 1000,
	})
}

export function riskQuery(id: number) {
	return queryOptions({
		queryKey: riskKeys.detail(id),
		queryFn: () => request<Risk>('get', `/risks/${id}`),
	})
}

export function riskHistoryQuery(id: number) {
	return queryOptions({
		queryKey: riskKeys.history(id),
		queryFn: async () => (await request<Paginated<RiskHistoryEntry>>('get', `/risks/${id}/history`, {query: {per_page: API_MAX_PER_PAGE}})).items,
	})
}

// Every page of what a filter selects, for an export. The server caps the page size at its own
// maximum, so the loop follows the total instead of trusting the size it asked for.
export const MAX_RISK_EXPORT = 5000

export async function fetchAllRisks(
	projectId: number | null,
	filters: RiskFilters,
	limit = MAX_RISK_EXPORT,
): Promise<{risks: Risk[], truncated: boolean, total: number}> {
	const risks: Risk[] = []
	let total = 0
	for (let page = 1; ; page++) {
		const result = await request<Paginated<Risk>>('get', listUrl(projectId), {
			query: riskListQuery({...filters, page, perPage: API_MAX_PER_PAGE}, projectId !== null),
		})
		total = result.total
		risks.push(...result.items)
		if (result.items.length === 0 || risks.length >= total || risks.length >= limit) {
			break
		}
	}
	return {risks: risks.slice(0, limit), truncated: total > limit, total}
}

// Every change can alter a list, a detail and the history, so mutations end by invalidating them.
// Cache writes happen only in the callbacks, with the client they are given.
function invalidateRisks(queryClient: QueryClient) {
	return queryClient.invalidateQueries({queryKey: riskKeys.all})
}

export function createRiskMutation() {
	return mutationOptions({
		mutationFn: ({projectId, ...body}: RiskInput & {projectId: number}) => request<Risk>('post', `/projects/${projectId}/risks`, {
			body,
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateRisks(context.client),
	})
}

export function updateRiskMutation() {
	return mutationOptions({
		mutationFn: ({id, ...body}: RiskInput & {id: number}) => request<Risk>('put', `/risks/${id}`, {
			body,
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateRisks(context.client),
	})
}

export function changeRiskStatusMutation() {
	return mutationOptions({
		mutationFn: ({id, status, note}: {id: number, status: RiskStatus, note?: string}) => request<Risk>('post', `/risks/${id}/status`, {
			body: {status, ...(note ? {note} : {})},
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateRisks(context.client),
	})
}

export function deleteRiskMutation() {
	return mutationOptions({
		mutationFn: (id: number) => request<void>('delete', `/risks/${id}`),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateRisks(context.client),
	})
}
