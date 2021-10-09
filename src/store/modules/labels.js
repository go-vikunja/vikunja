import LabelService from '@/services/label'
import {setLoading} from '@/store/helper'

/**
 * Returns the labels by id if found
 * @param {Object} state 
 * @param {Array} ids 
 * @returns {Array}
 */
function getLabelsByIds(state, ids) {
	return Object.values(state.labels).filter(({id}) => ids.includes(id))
}

/**
 * Checks if a list of labels is available in the store and filters them then query
 * @param {Object} state 
 * @param {Array} labels 
 * @param {String} query 
 * @returns {Array}
 */
 function filterLabelsByQuery(state, labels, query) {
	const labelIds = labels.map(({id}) => id)
	const foundLabels = getLabelsByIds(state, labelIds)
	const labelQuery = query.toLowerCase()

	return foundLabels.filter(({title}) => {
		return !title.toLowerCase().includes(labelQuery)
	})
}

export default {
	namespaced: true,
	state: () => ({
		// The labels are stored as an object which has the label ids as keys.
		labels: {},
		loaded: false,
	}),
	mutations: {
		setLabels(state, labels) {
			labels.forEach(l => {
				state.labels[l.id] = l
			})
		},
		setLabel(state, label) {
			state.labels[label.id] = label
		},
		removeLabelById(state, label) {
			delete state.labels[label.id]
		},
		setLoaded(state, loaded) {
			state.loaded = loaded
		},
	},
	getters: {
		getLabelsByIds(state) {
			return (ids) => getLabelsByIds(state, ids)
		},
		filterLabelsByQuery(state) {
			return (...arr) => filterLabelsByQuery(state, ...arr)
		},
	},
	actions: {
		loadAllLabels(ctx, {forceLoad} = {}) {
			if (ctx.state.loaded && !forceLoad) {
				return Promise.resolve()
			}

			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			const getAllLabels = (page = 1) => {
				return labelService.getAll({}, {}, page)
					.then(labels => {
						if (page < labelService.totalPages) {
							return getAllLabels(page + 1)
								.then(nextLabels => {
									return labels.concat(nextLabels)
								})
						} else {
							return labels
						}
					})
					.catch(e => {
						return Promise.reject(e)
					})
			}

			return getAllLabels()
				.then(r => {
					ctx.commit('setLabels', r)
					ctx.commit('setLoaded', true)
					return Promise.resolve(r)
				})
				.finally(() => cancel())
		},
		deleteLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			return labelService.delete(label)
				.then(r => {
					ctx.commit('removeLabelById', label)
					return Promise.resolve(r)
				})
				.finally(() => cancel())
		},
		updateLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			return labelService.update(label)
				.then(r => {
					ctx.commit('setLabel', r)
					return Promise.resolve(r)
				})
				.finally(() => cancel())
		},
		createLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			return labelService.create(label)
				.then(r => {
					ctx.commit('setLabel', r)
					return Promise.resolve(r)
				})
				.finally(() => cancel())
		},
	},
}
