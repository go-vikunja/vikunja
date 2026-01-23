import {describe, it, expect, beforeEach} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {useViewFiltersStore} from './viewFilters'

describe('viewFilters store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('should store and retrieve query params for a view', () => {
		const store = useViewFiltersStore()

		store.setViewQuery(18, {dateFrom: '2026-01-01', dateTo: '2026-01-31'})

		expect(store.getViewQuery(18)).toEqual({dateFrom: '2026-01-01', dateTo: '2026-01-31'})
	})

	it('should return empty object for views without stored params', () => {
		const store = useViewFiltersStore()

		expect(store.getViewQuery(999)).toEqual({})
	})

	it('should clear query params for a view', () => {
		const store = useViewFiltersStore()

		store.setViewQuery(18, {dateFrom: '2026-01-01', dateTo: '2026-01-31'})
		store.clearViewQuery(18)

		expect(store.getViewQuery(18)).toEqual({})
	})

	it('should update existing query params', () => {
		const store = useViewFiltersStore()

		store.setViewQuery(18, {dateFrom: '2026-01-01', dateTo: '2026-01-31'})
		store.setViewQuery(18, {dateFrom: '2026-02-01', dateTo: '2026-02-28'})

		expect(store.getViewQuery(18)).toEqual({dateFrom: '2026-02-01', dateTo: '2026-02-28'})
	})
})
