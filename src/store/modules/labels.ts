import type { Module } from 'vuex'

import {i18n} from '@/i18n'
import {success} from '@/message'
import LabelService from '@/services/label'
import {setLoading} from '@/store/helper'
import type { LabelState, RootStoreState } from '@/store/types'
import {getLabelsByIds, filterLabelsByQuery} from '@/helpers/labels'
import {createNewIndexer} from '@/indexes'
import type { ILabel } from '@/modelTypes/ILabel'

const {add, remove, update} = createNewIndexer('labels', ['title', 'description'])

async function getAllLabels(page = 1): Promise<ILabel[]> {
	const labelService = new LabelService()
	const labels  = await labelService.getAll({}, {}, page) as ILabel[]
	if (page < labelService.totalPages) {
		const nextLabels = await getAllLabels(page + 1)
		return labels.concat(nextLabels)
	} else {
		return labels
	}
}

const LabelStore : Module<LabelState, RootStoreState> = {
	namespaced: true,
	state: () => ({
		labels: {},
		loaded: false,
	}),
	mutations: {
		setLabels(state, labels: ILabel[]) {
			labels.forEach(l => {
				state.labels[l.id] = l
				add(l)
			})
		},
		setLabel(state, label: ILabel) {
			state.labels[label.id] = label
			update(label)
		},
		removeLabelById(state, label: ILabel) {
			remove(label)
			delete state.labels[label.id]
		},
		setLoaded(state, loaded: boolean) {
			state.loaded = loaded
		},
	},
	getters: {
		getLabelsByIds(state) {
			return (ids: ILabel['id'][]) => getLabelsByIds(state, ids)
		},
		filterLabelsByQuery(state) {
			return (labelsToHide: ILabel[], query: string) => filterLabelsByQuery(state, labelsToHide, query)
		},
		getLabelsByExactTitles(state) {
			return labelTitles => Object
				.values(state.labels)
				.filter(({title}) => labelTitles.some(l => l.toLowerCase() === title.toLowerCase()))
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
		async deleteLabel(ctx, label: ILabel) {
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
		async updateLabel(ctx, label: ILabel) {
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
		async createLabel(ctx, label: ILabel) {
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

export default LabelStore