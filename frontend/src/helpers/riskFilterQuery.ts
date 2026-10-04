import type {LocationQuery, LocationQueryRaw} from 'vue-router'

import type {RiskFilters, RiskSortKey} from '@/client/queries/risks'
import {isRiskRating, isRiskStatus, RISK_STATUSES} from '@/helpers/riskRating'
import type {RiskRating, RiskStatus} from '@/helpers/riskRating'

// The filters of a risk page live in its URL, so a filtered view can be bookmarked, shared and
// printed. Repeated values are repeated query parameters, never a comma list: a category may
// contain a comma.

// Without a status in the URL the list shows what is still being worked on. Closed risks come
// back with "Show closed", or by choosing the status.
export const DEFAULT_RISK_STATUSES: RiskStatus[] = ['open', 'mitigating', 'accepted']

const SORT_KEYS: RiskSortKey[] = ['score', 'probability', 'impact', 'due_date', 'status', 'title', 'created', 'updated', 'id']

function all(value: LocationQuery[string] | undefined): string[] {
	if (value === undefined || value === null) {
		return []
	}
	return (Array.isArray(value) ? value : [value]).filter((v): v is string => typeof v === 'string')
}

function first(value: LocationQuery[string] | undefined): string | undefined {
	return all(value)[0]
}

function ids(value: LocationQuery[string] | undefined, allowZero: boolean): number[] {
	const out: number[] = []
	for (const raw of all(value)) {
		const n = Number(raw)
		if (Number.isInteger(n) && (n > 0 || (allowZero && n === 0)) && !out.includes(n)) {
			out.push(n)
		}
	}
	return out
}

function sameSet<T>(a: T[], b: T[]): boolean {
	return a.length === b.length && a.every(x => b.includes(x))
}

// Unknown values are dropped, so a stale or hand-edited URL never sends a request the server refuses.
export function queryToRiskFilters(query: LocationQuery): RiskFilters {
	const filters: RiskFilters = {}

	const q = first(query.q)?.trim()
	if (q) filters.q = q

	const projectIds = ids(query.project, false)
	if (projectIds.length) filters.projectIds = projectIds

	const statuses = [...new Set(all(query.status).filter(isRiskStatus))]
	if (statuses.length) filters.statuses = statuses

	const ratings = [...new Set(all(query.rating).filter(isRiskRating))]
	if (ratings.length) filters.ratings = ratings

	const ownerIds = ids(query.owner, true)
	if (ownerIds.length) filters.ownerIds = ownerIds

	const categories = [...new Set(all(query.category).map(c => c.trim()).filter(Boolean))]
	if (categories.length) filters.categories = categories

	if (first(query.overdue) === '1') filters.overdue = true
	if (first(query.archived) === '1') filters.includeArchived = true

	const sort = first(query.sort)
	if (sort && (SORT_KEYS as string[]).includes(sort)) {
		filters.sortBy = sort as RiskSortKey
		filters.orderBy = first(query.order) === 'desc' ? 'desc' : 'asc'
	}

	return filters
}

export function riskFiltersToQuery(filters: RiskFilters): LocationQueryRaw {
	const query: LocationQueryRaw = {}

	if (filters.q) query.q = filters.q
	if (filters.projectIds?.length) query.project = filters.projectIds.map(String)
	// The default selection is not written, so the plain URL stays the default view.
	if (filters.statuses?.length && !sameSet(filters.statuses, DEFAULT_RISK_STATUSES)) {
		query.status = filters.statuses
	}
	if (filters.ratings?.length) query.rating = filters.ratings
	if (filters.ownerIds?.length) query.owner = filters.ownerIds.map(String)
	if (filters.categories?.length) query.category = filters.categories
	if (filters.overdue) query.overdue = '1'
	if (filters.includeArchived) query.archived = '1'
	if (filters.sortBy) {
		query.sort = filters.sortBy
		query.order = filters.orderBy ?? 'asc'
	}
	return query
}

// What the list actually asks for: the default statuses where none were chosen.
export function effectiveRiskStatuses(filters: RiskFilters): RiskStatus[] {
	return filters.statuses?.length ? filters.statuses : DEFAULT_RISK_STATUSES
}

export function withEffectiveStatuses(filters: RiskFilters): RiskFilters {
	return {...filters, statuses: effectiveRiskStatuses(filters)}
}

export function showsClosed(filters: RiskFilters): boolean {
	return effectiveRiskStatuses(filters).includes('closed')
}

// Turns closed risks on or off without touching the other statuses that were chosen.
export function toggleShowClosed(filters: RiskFilters, show: boolean): RiskFilters {
	const current = effectiveRiskStatuses(filters)
	const next = show
		? [...new Set<RiskStatus>([...current, 'closed'])]
		: current.filter(s => s !== 'closed')
	// Hiding closed risks while only closed was chosen would leave nothing: back to the default.
	return {...filters, statuses: next.length ? RISK_STATUSES.filter(s => next.includes(s)) : undefined}
}

export interface RiskFilterLabels {
	status: (status: RiskStatus) => string
	rating: (rating: RiskRating) => string
	search: (q: string) => string
	category: (category: string) => string
	overdue: string
	unowned: string
	owners: string
}

// A line of words for the head of a printout, so the paper says what the list was narrowed by. The
// statuses are always named, also when they are the default ones.
export function describeRiskFilters(filters: RiskFilters, labels: RiskFilterLabels): string {
	const parts: string[] = [effectiveRiskStatuses(filters).map(labels.status).join(', ')]
	if (filters.ratings?.length) parts.push(filters.ratings.map(labels.rating).join(', '))
	if (filters.ownerIds?.length) {
		const unowned = filters.ownerIds.includes(0)
		const others = filters.ownerIds.some(id => id !== 0)
		parts.push([unowned ? labels.unowned : '', others ? labels.owners : ''].filter(Boolean).join(', '))
	}
	if (filters.categories?.length) parts.push(filters.categories.map(labels.category).join(', '))
	if (filters.q) parts.push(labels.search(filters.q))
	if (filters.overdue) parts.push(labels.overdue)
	return parts.join('; ')
}

// How many filters narrow the list, for the badge on a collapsed filter bar. The default
// statuses do not count.
export function activeRiskFilterCount(filters: RiskFilters): number {
	let count = 0
	if (filters.q) count++
	if (filters.projectIds?.length) count++
	if (filters.statuses?.length && !sameSet(filters.statuses, DEFAULT_RISK_STATUSES)) count++
	if (filters.ratings?.length) count++
	if (filters.ownerIds?.length) count++
	if (filters.categories?.length) count++
	if (filters.overdue) count++
	if (filters.includeArchived) count++
	return count
}
