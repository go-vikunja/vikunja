import {beforeEach, describe, expect, it} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {useConfigStore} from './config'

describe('config store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('should have an empty apiBase initially', () => {
		const store = useConfigStore()
		expect(store.apiBase).toBe('')
	})

	it('should update apiBase when setApiUrl is called', () => {
		const store = useConfigStore()
		const newUrl = 'http://localhost:3456/api/v1'
		store.setApiUrl(newUrl)
		expect(store.apiBase).toBe(newUrl)
	})

	it('should clean up the url when setApiUrl is called', () => {
		const store = useConfigStore()
		const newUrl = 'http://localhost:3456/api/v1/'
		store.setApiUrl(newUrl)
		expect(store.apiBase).toBe('http://localhost:3456/api/v1')
	})
})