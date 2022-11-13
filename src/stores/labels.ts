import {computed, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

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
	[id: ILabel['id']]: ILabel
}

export const useLabelStore = defineStore('label', () => {
	// The labels are stored as an object which has the label ids as keys.
	const labels = ref<LabelState>({})
	const isLoading = ref(false)

	const getLabelsByIds = computed(() => {
		return (ids: ILabel['id'][]) => Object.values(labels.value).filter(({id}) => ids.includes(id))
	})

	// **
	// * Checks if a project of labels is available in the store and filters them then query
	// **
	const filterLabelsByQuery = computed(() => {
		return (labelsToHide: ILabel[], query: string) => {
			const labelIdsToHide: number[] = labelsToHide.map(({id}) => id)
		
			return search(query)
					?.filter(value => !labelIdsToHide.includes(value))
					.map(id => labels.value[id])
				|| []
		}
	})

	const getLabelsByExactTitles = computed(() => {
		return (labelTitles: string[]) => Object
			.values(labels.value)
			.filter(({title}) => labelTitles.some(l => l.toLowerCase() === title.toLowerCase()))
	})


	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setLabels(newLabels: ILabel[]) {
		newLabels.forEach(l => {
			labels.value[l.id] = l
			add(l)
		})
	}

	function setLabel(label: ILabel) {
		labels.value[label.id] = label
		update(label)
	}

	function removeLabelById(label: ILabel) {
		remove(label)
		delete labels.value[label.id]
	}

	async function loadAllLabels({forceLoad} : {forceLoad?: boolean} = {}) {
		if (isLoading.value && !forceLoad) {
			return
		}

		const cancel = setModuleLoading(setIsLoading)

		try {
			const newLabels = await getAllLabels()
			setLabels(newLabels)
			return newLabels
		} finally {
			cancel()
		}
	}

	async function deleteLabel(label: ILabel) {
		const cancel = setModuleLoading(setIsLoading)
		const labelService = new LabelService()

		try {
			const result = await labelService.delete(label)
			removeLabelById(label)
			success({message: i18n.global.t('label.deleteSuccess')})
			return result
		} finally {
			cancel()
		}
	}

	async function updateLabel(label: ILabel) {
		const cancel = setModuleLoading(setIsLoading)
		const labelService = new LabelService()

		try {
			const newLabel = await labelService.update(label)
			setLabel(newLabel)
			success({message: i18n.global.t('label.edit.success')})
			return newLabel
		} finally {
			cancel()
		}
	}

	async function createLabel(label: ILabel) {
		const cancel = setModuleLoading(setIsLoading)
		const labelService = new LabelService()

		try {
			const newLabel = await labelService.create(label) as ILabel
			setLabel(newLabel)
			return newLabel
		} finally {
			cancel()
		}
	}

	return {
		labels,
		isLoading,

		getLabelsByIds,
		filterLabelsByQuery,
		getLabelsByExactTitles,

		setLabels,
		setLabel,
		removeLabelById,
		loadAllLabels,
		deleteLabel,
		updateLabel,
		createLabel,
		
	}
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useLabelStore, import.meta.hot))
}