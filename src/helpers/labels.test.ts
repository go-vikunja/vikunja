import {describe, it, expect} from 'vitest'

import {filterLabelsByQuery} from './labels'
import {createNewIndexer} from '../indexes'

const {add} = createNewIndexer('labels', ['title', 'description'])

describe('filter labels', () => {
	const state = {
		labels: {
			1: {id: 1, title: 'label1'},
			2: {id: 2, title: 'label2'},
			3: {id: 3, title: 'label3'},
			4: {id: 4, title: 'label4'},
			5: {id: 5, title: 'label5'},
			6: {id: 6, title: 'label6'},
			7: {id: 7, title: 'label7'},
			8: {id: 8, title: 'label8'},
			9: {id: 9, title: 'label9'},
		},
	}

	Object.values(state.labels).forEach(add)

	it('should return an empty array for an empty query', () => {
		const labels = filterLabelsByQuery(state, [], '')

		expect(labels).toHaveLength(0)
	})
	it('should return labels for a query', () => {
		const labels = filterLabelsByQuery(state, [], 'label2')

		expect(labels).toHaveLength(1)
		expect(labels[0].title).toBe('label2')
	})
	it('should not return found but hidden labels', () => {
		interface label {
			id: number,
			title: string,
		}

		const labelsToHide: label[] = [{id: 1, title: 'label1'}]
		const labels = filterLabelsByQuery(state, labelsToHide, 'label1')

		expect(labels).toHaveLength(0)
	})
})
