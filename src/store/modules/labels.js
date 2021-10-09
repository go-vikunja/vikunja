import LabelService from '@/services/label'
import {setLoading} from '@/store/helper'
import { success } from '@/message'
import {i18n} from '@/i18n'

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

const labelService = new LabelService()

const getAllLabels = async (page = 1) => {
	const labels = await labelService.getAll({}, {}, page)
	if (page < labelService.totalPages) {
		const nextLabels = await getAllLabels(page + 1)
		return labels.concat(nextLabels)
	} else {
		return labels
	}
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
		async loadAllLabels(ctx, {forceLoad} = {}) {
			if (ctx.state.loaded && !forceLoad) {
				return
			}

			const cancel = setLoading(ctx, 'labels')

			try {
				const labels = await getAllLabels()
				ctx.commit('setLabels', labels)
				ctx.commit('setLoaded', true)
				return labels
			} finally {
				cancel()
			}
		},
		async deleteLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			try {
				const result = await labelService.delete(label)
				ctx.commit('removeLabelById', label)
				success({message: i18n.global.t('label.deleteSuccess')})
				return result
			} finally {
				cancel()
			}
		},
		async updateLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			try {
				const newLabel = await labelService.update(label)
				ctx.commit('setLabel', newLabel)
				success({message: i18n.global.t('label.edit.success')})
				return newLabel
			} finally {
				cancel()
			}
		},
		async createLabel(ctx, label) {
			const cancel = setLoading(ctx, 'labels')
			const labelService = new LabelService()

			try {
				const newLabel = await labelService.create(label)
				ctx.commit('setLabel', newLabel)
				return newLabel
			} finally {
				cancel()
			}
		},
	},
}
