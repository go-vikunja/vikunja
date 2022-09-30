import { acceptHMRUpdate, defineStore } from 'pinia'

import LabelService from '@/services/label'
import {success} from '@/message'
import {i18n} from '@/i18n'
import {createNewIndexer} from '@/indexes'
import {setModuleLoading} from '@/stores/helper'
import type {ILabel} from '@/modelTypes/ILabel'

const {add, remove, update, search} = createNewIndexer('labels', ['title', 'description'])

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

export interface LabelState {
	labels: {
		[id: ILabel['id']]: ILabel
	},
	isLoading: boolean,
}

export const useLabelStore = defineStore('label', {
	state: () : LabelState => ({
		// The labels are stored as an object which has the label ids as keys.
		labels: {},
		isLoading: false,
	}),
	
	getters: {
		getLabelsByIds(state) {
			return (ids: ILabel['id'][]) => Object.values(state.labels).filter(({id}) => ids.includes(id))
		},
		// **
		// * Checks if a list of labels is available in the store and filters them then query
		// **
		filterLabelsByQuery(state) {
			return (labelsToHide: ILabel[], query: string) => {
				const labelIdsToHide: number[] = labelsToHide.map(({id}) => id)
			
				return search(query)
						?.filter(value => !labelIdsToHide.includes(value))
						.map(id => state.labels[id])
					|| []
			}
		},
		getLabelsByExactTitles(state) {
			return (labelTitles: string[]) => Object
				.values(state.labels)
				.filter(({title}) => labelTitles.some(l => l.toLowerCase() === title.toLowerCase()))
		},
	},

	actions: {
		setIsLoading(isLoading: boolean) {
			this.isLoading = isLoading
		},

		setLabels(labels: ILabel[]) {
			labels.forEach(l => {
				this.labels[l.id] = l
				add(l)
			})
		},

		setLabel(label: ILabel) {
			this.labels[label.id] = label
			update(label)
		},

		removeLabelById(label: ILabel) {
			remove(label)
			delete this.labels[label.id]
		},

		async loadAllLabels({forceLoad} : {forceLoad?: boolean} = {}) {
			if (this.isLoading && !forceLoad) {
				return
			}

			const cancel = setModuleLoading(this)

			try {
				const labels = await getAllLabels()
				this.setLabels(labels)
				return labels
			} finally {
				cancel()
			}
		},

		async deleteLabel(label: ILabel) {
			const cancel = setModuleLoading(this)
			const labelService = new LabelService()

			try {
				const result = await labelService.delete(label)
				this.removeLabelById(label)
				success({message: i18n.global.t('label.deleteSuccess')})
				return result
			} finally {
				cancel()
			}
		},

		async updateLabel(label: ILabel) {
			const cancel = setModuleLoading(this)
			const labelService = new LabelService()

			try {
				const newLabel = await labelService.update(label)
				this.setLabel(newLabel)
				success({message: i18n.global.t('label.edit.success')})
				return newLabel
			} finally {
				cancel()
			}
		},

		async createLabel(label: ILabel) {
			const cancel = setModuleLoading(this)
			const labelService = new LabelService()

			try {
				const newLabel = await labelService.create(label) as ILabel
				this.setLabel(newLabel)
				return newLabel
			} finally {
				cancel()
			}
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useLabelStore, import.meta.hot))
}