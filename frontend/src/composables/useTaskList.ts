import {ref, shallowReactive, watch, computed, type ComputedGetter} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import TaskCollectionService, {
	type ExpandTaskFilterParam,
	getDefaultTaskFilterParams,
	type TaskFilterParams,
} from '@/services/taskCollection'
import type {ITask} from '@/modelTypes/ITask'
import {error} from '@/message'
import type {IProject} from '@/modelTypes/IProject'
import {useAuthStore} from '@/stores/auth'
import type {IProjectView} from '@/modelTypes/IProjectView'

export type Order = 'asc' | 'desc' | 'none'

export interface SortBy {
	id?: Order
	index?: Order
	done?: Order
	title?: Order
	priority?: Order
	due_date?: Order
	start_date?: Order
	end_date?: Order
	percent_done?: Order
	created?: Order
	updated?: Order
	done_at?: Order,
}

const SORT_BY_DEFAULT: SortBy = {
	id: 'desc',
}

// This makes sure an id sort order is always sorted last.
// When tasks would be sorted first by id and then by whatever else was specified, the id sort takes
// precedence over everything else, making any other sort columns pretty useless.
function formatSortOrder(sortBy, params) {
	let hasIdFilter = false
	const sortKeys = Object.keys(sortBy)
	for (const s of sortKeys) {
		if (s === 'id') {
			sortKeys.splice(s, 1)
			hasIdFilter = true
			break
		}
	}
	if (hasIdFilter) {
		sortKeys.push('id')
	}
	params.sort_by = sortKeys
	params.order_by = sortKeys.map(s => sortBy[s])

	return params
}

/**
 * This mixin provides a base set of methods and properties to get tasks.
 */
export function useTaskList(
	projectIdGetter: ComputedGetter<IProject['id']>,
	projectViewIdGetter: ComputedGetter<IProjectView['id']>,
	sortByDefault: SortBy = SORT_BY_DEFAULT,
	expandGetter: ComputedGetter<ExpandTaskFilterParam> = () => 'subtasks',
) {
	
	const projectId = computed(() => projectIdGetter())
	const projectViewId = computed(() => projectViewIdGetter())

	const params = ref<TaskFilterParams>({...getDefaultTaskFilterParams()})

	const page = useRouteQuery('page', '1', { transform: Number })
	const filter = useRouteQuery('filter')
	const s = useRouteQuery('s')

	watch(filter, v => { params.value.filter = v ?? '' }, { immediate: true })
	watch(s, v => { params.value.s = v ?? '' }, { immediate: true })

	watch(() => params.value.filter, v => { filter.value = v || undefined })
	watch(() => params.value.s, v => { s.value = v || undefined })

	const sortBy = ref({ ...sortByDefault })
	
	const allParams = computed(() => {
		const loadParams = {...params.value}

		return formatSortOrder(sortBy.value, loadParams)
	})

	watch(
		[params, sortBy, page],
		([, , newPage], [, , oldPage]) => {
			if (newPage === oldPage) {
				page.value = 1
			}
		},
		{deep: true},
	)
	
	const authStore = useAuthStore()
	
	const getAllTasksParams = computed(() => {
		return [
			{
				projectId: projectId.value,
				viewId: projectViewId.value,
			},
			{
				...allParams.value,
				filter_timezone: authStore.settings.timezone,
				expand: expandGetter(),
			},
			page.value,
		]
	})

	const taskCollectionService = shallowReactive(new TaskCollectionService())
	const loading = computed(() => taskCollectionService.loading)
	const totalPages = computed(() => taskCollectionService.totalPages)

	const tasks = ref<ITask[]>([])
	async function loadTasks(resetBeforeLoad: boolean = true) {
		if(resetBeforeLoad) {
			tasks.value = []
		}
		try {
			tasks.value = await taskCollectionService.getAll(...getAllTasksParams.value)
		} catch (e) {
			error(e)
		}
		return tasks.value
	}

	// Only listen for query path changes
	watch(() => JSON.stringify(getAllTasksParams.value), (newParams, oldParams) => {
		if (oldParams === newParams) {
			return
		}

		loadTasks()
	}, { immediate: true })

	return {
		tasks,
		loading,
		totalPages,
		currentPage: page,
		loadTasks,
		params,
		sortByParam: sortBy,
	}
}
