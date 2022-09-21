import {setActivePinia, createPinia} from 'pinia'
import {describe, it, expect, beforeEach} from 'vitest'

import {useLabelStore} from './labels'

import type { ILabel } from '@/modelTypes/ILabel'

const MOCK_LABELS = {
	1: {id: 1, title: 'label1'},
	2: {id: 2, title: 'label2'},
	3: {id: 3, title: 'label3'},
	4: {id: 4, title: 'label4'},
	5: {id: 5, title: 'label5'},
	6: {id: 6, title: 'label6'},
	7: {id: 7, title: 'label7'},
	8: {id: 8, title: 'label8'},
	9: {id: 9, title: 'label9'},
}

function setupStore() {
	const store = useLabelStore()
	store.setLabels(Object.values(MOCK_LABELS) as ILabel[])
	return store
}

describe('filter labels', () => {
  beforeEach(() => {
    // creates a fresh pinia and make it active so it's automatically picked
    // up by any useStore() call without having to pass it to it:
    // `useStore(pinia)`
    setActivePinia(createPinia())
  })

	it('should return an empty array for an empty query', () => {
		const store = setupStore()
		const labels = store.filterLabelsByQuery([], '')

		expect(labels).toHaveLength(0)
	})
	it('should return labels for a query', () => {
		const store = setupStore()
		const labels = store.filterLabelsByQuery([], 'label2')

		expect(labels).toHaveLength(1)
		expect(labels[0].title).toBe('label2')
	})
	it('should not return found but hidden labels', () => {
		const store = setupStore()

		const labelsToHide = [{id: 1, title: 'label1'}] as ILabel[]
		const labels = store.filterLabelsByQuery(labelsToHide, 'label1')

		expect(labels).toHaveLength(0)
	})
})
