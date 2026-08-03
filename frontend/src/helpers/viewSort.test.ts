import {describe, expect, it} from 'vitest'

import {
	decodeSortSelection,
	encodeSortSelection,
	sortByFromViewFilter,
	viewFilterSortFromSortBy,
} from './viewSort'

describe('sortByFromViewFilter', () => {
	it('returns null when there is no sort', () => {
		expect(sortByFromViewFilter(undefined)).toBeNull()
		expect(sortByFromViewFilter({
			sort_by: [],
			order_by: [],
			filter: '',
			filter_include_nulls: false,
			s: '',
		})).toBeNull()
	})

	it('reads snake_case sort arrays from the API', () => {
		expect(sortByFromViewFilter({
			sort_by: ['due_date', 'id'],
			order_by: ['asc', 'asc'],
			filter: '',
			filter_include_nulls: false,
			s: '',
		})).toEqual({due_date: 'asc'})
	})

	it('reads camelCase sort arrays from assignData', () => {
		expect(sortByFromViewFilter({
			sortBy: ['priority'],
			orderBy: ['desc'],
			filter: '',
			filter_include_nulls: false,
			s: '',
		} as never)).toEqual({priority: 'desc'})
	})
})

describe('viewFilterSortFromSortBy / encodeSortSelection', () => {
	it('round-trips a primary sort selection', () => {
		const sort = decodeSortSelection('due_date:asc')
		expect(sort).toEqual({due_date: 'asc'})
		expect(viewFilterSortFromSortBy(sort)).toEqual({
			sort_by: ['due_date'],
			order_by: ['asc'],
		})
		expect(encodeSortSelection(sort)).toBe('due_date:asc')
	})

	it('treats position as manual', () => {
		expect(encodeSortSelection({position: 'asc'})).toBe('position:asc')
		expect(encodeSortSelection(null)).toBe('position:asc')
	})
})
