interface label {
	id: number,
	title: string,
}

interface labelState {
	labels: label[],
}

/**
 * Checks if a list of labels is available in the store and filters them then query
 * @param {Object} state
 * @param {Array} labelsToHide
 * @param {String} query
 * @returns {Array}
 */
export function filterLabelsByQuery(state: labelState, labelsToHide: label[], query: string) {
	if (query === '') {
		return []
	}

	const labelQuery = query.toLowerCase()
	const labelIds = labelsToHide.map(({id}) => id)
	return Object
		.values(state.labels)
		.filter(({id, title}) => {
			return !labelIds.includes(id) && title.toLowerCase().includes(labelQuery)
		})
}
