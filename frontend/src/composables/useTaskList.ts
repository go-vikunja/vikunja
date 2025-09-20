import {ref, shallowReactive, watch, computed, type ComputedGetter} from 'vue'
import {useRouteQuery} from '@vueuse/router'

import TaskCollectionService, {
	type ExpandTaskFilterParam,
	getDefaultTaskFilterParams,
	type TaskFilterParams,
} from '@/services/taskCollection'
import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'
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
function formatSortOrder(sortBy: SortBy, params: TaskFilterParams) {
	let hasIdFilter = false
	const sortKeys = Object.keys(sortBy)
	for (let i = 0; i < sortKeys.length; i++) {
		if (sortKeys[i] === 'id') {
			sortKeys.splice(i, 1)
			hasIdFilter = true
			break
		}
	}
	if (hasIdFilter) {
		sortKeys.push('id')
	}
	params.sort_by = sortKeys.filter(key =>
		['start_date', 'end_date', 'due_date', 'done', 'id', 'position', 'title'].includes(key),
	) as ('start_date' | 'end_date' | 'due_date' | 'done' | 'id' | 'position' | 'title')[]
	params.order_by = sortKeys.map(s => sortBy[s as keyof SortBy]).filter(Boolean) as ('asc' | 'desc')[]

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

	watch(filter, v => { params.value.filter = Array.isArray(v) ? v[0] ?? '' : v ?? '' }, { immediate: true })
	watch(s, v => { params.value.s = Array.isArray(v) ? v[0] ?? '' : v ?? '' }, { immediate: true })

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
			undefined as ITask | undefined,
			{
				...allParams.value,
				filter_timezone: authStore.settings.timezone,
				expand: expandGetter(),
				projectId: projectId.value,
				viewId: projectViewId.value,
			},
			page.value,
		] as const
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
			const result = await taskCollectionService.getAll(...getAllTasksParams.value)
			// Filter out buckets, only keep tasks
			tasks.value = result.filter((item): item is ITask =>
				!('project_view_id' in item) && !('projectViewId' in item),
			)
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
