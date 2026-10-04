import {describe, expect, it} from 'vitest'

import {
	activeRiskFilterCount,
	DEFAULT_RISK_STATUSES,
	describeRiskFilters,
	type RiskFilterLabels,
	effectiveRiskStatuses,
	queryToRiskFilters,
	riskFiltersToQuery,
	showsClosed,
	toggleShowClosed,
	withEffectiveStatuses,
} from './riskFilterQuery'

describe('queryToRiskFilters', () => {
	it('reads every filter, single and repeated values', () => {
		const filters = queryToRiskFilters({
			q: ' supplier ',
			project: ['3', '5'],
			status: ['open', 'closed'],
			rating: 'critical',
			owner: ['0', '7'],
			category: ['Schedule', 'Cost, schedule'],
			overdue: '1',
			archived: '1',
			sort: 'score',
			order: 'desc',
		})

		expect(filters).toEqual({
			q: 'supplier',
			projectIds: [3, 5],
			statuses: ['open', 'closed'],
			ratings: ['critical'],
			ownerIds: [0, 7],
			categories: ['Schedule', 'Cost, schedule'],
			overdue: true,
			includeArchived: true,
			sortBy: 'score',
			orderBy: 'desc',
		})
	})

	it('an empty query is no filter at all', () => {
		expect(queryToRiskFilters({})).toEqual({})
	})

	it('drops values the server would refuse or that make no sense', () => {
		const filters = queryToRiskFilters({
			status: ['open', 'finished', 'open'],
			rating: ['huge', 'low'],
			project: ['abc', '-3', '0', '2.5', '4', '4'],
			owner: ['x', '-1', '9'],
			category: ['', '  '],
			sort: 'password',
			overdue: 'yes',
		})

		expect(filters).toEqual({statuses: ['open'], ratings: ['low'], projectIds: [4], ownerIds: [9]})
	})

	it('a sort without an order is ascending, nothing but desc means asc', () => {
		expect(queryToRiskFilters({sort: 'title'})).toEqual({sortBy: 'title', orderBy: 'asc'})
		expect(queryToRiskFilters({sort: 'title', order: 'sideways'}).orderBy).toBe('asc')
	})

	it('ignores null entries of the router', () => {
		expect(queryToRiskFilters({status: [null, 'open'], q: null})).toEqual({statuses: ['open']})
	})
})

describe('riskFiltersToQuery', () => {
	it('round-trips', () => {
		const filters = {
			q: 'delay', projectIds: [1, 2], statuses: ['mitigating' as const], ratings: ['high' as const, 'low' as const],
			ownerIds: [0, 4], categories: ['A, B'], overdue: true, includeArchived: true, sortBy: 'due_date' as const, orderBy: 'desc' as const,
		}

		expect(queryToRiskFilters(riskFiltersToQuery(filters) as never)).toEqual(filters)
	})

	it('does not write the default statuses, so the plain URL is the default view', () => {
		expect(riskFiltersToQuery({statuses: [...DEFAULT_RISK_STATUSES]})).toEqual({})
		expect(riskFiltersToQuery({statuses: ['accepted', 'open', 'mitigating']})).toEqual({})
		expect(riskFiltersToQuery({statuses: ['open', 'closed']})).toEqual({status: ['open', 'closed']})
	})

	it('leaves out what is empty or off', () => {
		expect(riskFiltersToQuery({q: '', projectIds: [], ratings: [], overdue: false, includeArchived: false})).toEqual({})
	})
})

describe('statuses', () => {
	it('the default hides closed risks', () => {
		expect(effectiveRiskStatuses({})).toEqual(['open', 'mitigating', 'accepted'])
		expect(showsClosed({})).toBe(false)
		expect(withEffectiveStatuses({q: 'x'}).statuses).toEqual(DEFAULT_RISK_STATUSES)
	})

	it('a chosen selection wins', () => {
		expect(effectiveRiskStatuses({statuses: ['closed']})).toEqual(['closed'])
		expect(showsClosed({statuses: ['closed']})).toBe(true)
	})

	it('show closed adds it to what is chosen, in a stable order', () => {
		expect(toggleShowClosed({}, true).statuses).toEqual(['open', 'mitigating', 'accepted', 'closed'])
		expect(toggleShowClosed({statuses: ['mitigating']}, true).statuses).toEqual(['mitigating', 'closed'])
	})

	it('hiding closed removes only closed, and falls back to the default when nothing is left', () => {
		expect(toggleShowClosed({statuses: ['open', 'closed']}, false).statuses).toEqual(['open'])
		expect(toggleShowClosed({statuses: ['closed']}, false).statuses).toBeUndefined()
		expect(toggleShowClosed({}, false).statuses).toEqual(DEFAULT_RISK_STATUSES)
	})

	it('keeps the other filters', () => {
		expect(toggleShowClosed({q: 'x', overdue: true}, true)).toMatchObject({q: 'x', overdue: true})
	})
})

describe('activeRiskFilterCount', () => {
	it('counts what narrows the list, not the default statuses or the sort', () => {
		expect(activeRiskFilterCount({})).toBe(0)
		expect(activeRiskFilterCount({statuses: DEFAULT_RISK_STATUSES, sortBy: 'title'})).toBe(0)
		expect(activeRiskFilterCount({q: 'a', ratings: ['low'], overdue: true})).toBe(3)
		expect(activeRiskFilterCount({statuses: ['closed'], projectIds: [1], ownerIds: [0], categories: ['x'], includeArchived: true})).toBe(5)
	})
})

describe('describeRiskFilters', () => {
	const labels: RiskFilterLabels = {
		status: s => s.toUpperCase(),
		rating: r => r.toUpperCase(),
		search: q => `search "${q}"`,
		category: c => `category ${c}`,
		overdue: 'overdue',
		unowned: 'no owner',
		owners: 'owners',
	}

	it('names the default statuses when nothing else is set', () => {
		expect(describeRiskFilters({}, labels)).toBe('OPEN, MITIGATING, ACCEPTED')
	})

	it('names every filter that is set', () => {
		expect(describeRiskFilters({
			statuses: ['open', 'closed'],
			ratings: ['high', 'critical'],
			ownerIds: [0, 4],
			categories: ['Schedule'],
			q: 'supplier',
			overdue: true,
		}, labels)).toBe('OPEN, CLOSED; HIGH, CRITICAL; no owner, owners; category Schedule; search "supplier"; overdue')
	})
})
