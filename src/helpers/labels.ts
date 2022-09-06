import {createNewIndexer} from '../indexes'

import type {LabelState} from '@/store/types'
import type {ILabel} from '@/models/label'

const {search} = createNewIndexer('labels', ['title', 'description'])

/**
 * Checks if a list of labels is available in the store and filters them then query
 * @param {Object} state
 * @param {Array} labelsToHide
 * @param {String} query
 * @returns {Array}
 */
export function filterLabelsByQuery(state: LabelState, labelsToHide: ILabel[], query: string) {
	const labelIdsToHide: number[] = labelsToHide.map(({id}) => id)

	return search(query)
			?.filter(value => !labelIdsToHide.includes(value))
			.map(id => state.labels[id])
		|| []
}


/**
 * Returns the labels by id if found
 * @param {Object} state
 * @param {Array} ids
 * @returns {Array}
 */
export function getLabelsByIds(state: LabelState, ids: ILabel['id'][]) {
	return Object.values(state.labels).filter(({id}) => ids.includes(id))
}