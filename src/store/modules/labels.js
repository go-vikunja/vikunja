import LabelService from '@/services/label'
import {setLoading} from '@/store/helper'
import {filterLabelsByQuery} from '@/helpers/labels'

/**
 * Returns the labels by id if found
 * @param {Object} state
 * @param {Array} ids
 * @returns {Array}
 */
function getLabelsByIds(state, ids) {
	return Object.values(state.labels).filter(({id}) => ids.includes(id))
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
			return (labelsToHide, query) => filterLabelsByQuery(state, labelsToHide, query)
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
				.catch(e => Promise.reject(e))
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
				.catch(e => Promise.reject(e))
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
				.catch(e => Promise.reject(e))
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
				.catch(e => Promise.reject(e))
				.finally(() => cancel())
		},
	},
}
