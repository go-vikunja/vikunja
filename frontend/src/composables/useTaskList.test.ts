import {describe, it, expect} from 'vitest'
import {buildStoredQuery, hasAnyTaskListQueryParam} from './useTaskList'

describe('buildStoredQuery', () => {
	it('includes sort when set', () => {
		expect(buildStoredQuery({sort: 'due_date:asc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'due_date:asc'})
	})

	it('includes filter and search when set', () => {
		expect(buildStoredQuery({sort: undefined, filter: 'done = false', s: 'foo', page: 1}))
			.toEqual({filter: 'done = false', s: 'foo'})
	})

	it('omits page when it equals the default of 1', () => {
		expect(buildStoredQuery({sort: 'id:desc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'id:desc'})
	})

	it('includes page when greater than 1', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 3}))
			.toEqual({page: '3'})
	})

	it('returns an empty object when nothing is set', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 1}))
			.toEqual({})
	})

	it('skips empty strings', () => {
		expect(buildStoredQuery({sort: '', filter: '', s: '', page: 1}))
			.toEqual({})
	})
})

describe('hasAnyTaskListQueryParam', () => {
	it('returns false when none of the four query params are present', () => {
		expect(hasAnyTaskListQueryParam({})).toBe(false)
		expect(hasAnyTaskListQueryParam({otherUnrelated: 'x'})).toBe(false)
	})

	it('returns true when sort is set', () => {
		expect(hasAnyTaskListQueryParam({sort: 'id:desc'})).toBe(true)
	})

	it('returns true when filter is set', () => {
		expect(hasAnyTaskListQueryParam({filter: 'done = false'})).toBe(true)
	})

	it('returns true when s is set', () => {
		expect(hasAnyTaskListQueryParam({s: 'foo'})).toBe(true)
	})

	it('returns true when page is set', () => {
		expect(hasAnyTaskListQueryParam({page: '2'})).toBe(true)
	})
})
