import {computed, ref, shallowReactive, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useRouter} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useDebounceFn} from '@vueuse/core'

import type {IProject} from '@/modelTypes/IProject'
import type {ISavedFilter} from '@/modelTypes/ISavedFilter'

import AbstractService from '@/services/abstractService'

import SavedFilterModel from '@/models/savedFilter'

import {useBaseStore} from '@/stores/base'
import {useProjectStore} from '@/stores/projects'

import {success} from '@/message'
import ProjectModel from '@/models/project'

/**
* Calculates the corresponding project id to this saved filter.
* This function matches the one in the api.
*/
function getProjectId(savedFilter: ISavedFilter) {
	let projectId = savedFilter.id * -1 - 1
	if (projectId > 0) {
		projectId = 0
	}
	return projectId
}

export function getSavedFilterIdFromProjectId(projectId: IProject['id']) {
	let filterId = projectId * -1 - 1
	// FilterIds from projectIds are always positive
	if (filterId < 0) {
		filterId = 0
	}
	return filterId
}

export function isSavedFilter(project: IProject) {
	return getSavedFilterIdFromProjectId(project?.id) > 0
}

export default class SavedFilterService extends AbstractService<ISavedFilter> {
	constructor() {
		super({
			get: '/filters/{id}',
			create: '/filters',
			update: '/filters/{id}',
			delete: '/filters/{id}',
		})
	}

	modelFactory(data) {
		return new SavedFilterModel(data)
	}
}

export function useSavedFilter(projectId?: MaybeRefOrGetter<IProject['id']>) {
	const router = useRouter()
	const {t} = useI18n({useScope:'global'})
	const projectStore = useProjectStore()

	const filterService = shallowReactive(new SavedFilterService())

	const filter = ref<ISavedFilter>(new SavedFilterModel())
	const filters = computed({
		get: () => filter.value.filters,
		set(value) {
			filter.value.filters = value
		},
	})

	// load SavedFilter
	watch(() => toValue(projectId), async (watchedProjectId) => {
		if (watchedProjectId === undefined) {
			return
		}

		// We assume the projectId in the route is the pseudoproject
		const savedFilterId = getSavedFilterIdFromProjectId(watchedProjectId)

		filter.value = await filterService.get(new SavedFilterModel({id: savedFilterId}))
		await validateTitleField()
	}, {immediate: true})

	async function createFilter() {
		filter.value = await filterService.create(filter.value)
		await projectStore.loadAllProjects()
		router.push({name: 'project.index', params: {projectId: getProjectId(filter.value)}})
	}

	async function saveFilter() {
		const response = await filterService.update(filter.value)
		await projectStore.loadAllProjects()
		success({message: t('filters.edit.success')})
		filter.value = response
		await useBaseStore().setCurrentProject(new ProjectModel({
			id: getProjectId(filter.value),
			title: filter.value.title,
		}))
		router.back()
	}

	async function deleteFilter() {	
		await filterService.delete(filter.value)
		await projectStore.loadAllProjects()
		success({message: t('filters.delete.success')})
		router.push({name: 'projects.index'})
	}

	const titleValid = ref(true)
	const validateTitleField = useDebounceFn(() => {
		titleValid.value = filter.value.title !== ''
	}, 100)

	async function createFilterWithValidation() {
		if (!titleValid.value) {
			return
		}
		return createFilter()
	}
	
	async function saveFilterWithValidation() {
		if (!titleValid.value) {
			return
		}
		return saveFilter()
	}

	return {
		createFilter,
		createFilterWithValidation,
		saveFilter,
		saveFilterWithValidation,
		deleteFilter,

		filter,
		filters,

		filterService,
		
		titleValid,
		validateTitleField,
	}
}
