import {ref, shallowReactive, watch, computed, type ComputedGetter} from 'vue'
import {useRouter} from 'vue-router'
import type {LocationQuery, LocationQueryRaw} from 'vue-router'
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
import {useViewFiltersStore} from '@/stores/viewFilters'
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
	position?: Order,
}

const VALID_SORT_FIELDS = new Set<string>(
	['id', 'index', 'done', 'title', 'priority', 'due_date', 'start_date',
		'end_date', 'percent_done', 'created', 'updated', 'done_at', 'position'],
)

function parseSortQuery(raw: string, fallback: SortBy): SortBy {
	const result: Record<string, Order> = {}
	for (const part of raw.split(',')) {
		const [field, order] = part.split(':')
		if (!VALID_SORT_FIELDS.has(field)) continue
		if (order !== 'asc' && order !== 'desc') continue
		result[field] = order
	}
	return Object.keys(result).length > 0 ? result as SortBy : {...fallback}
}

function serializeSortBy(sortBy: SortBy, defaultSort: SortBy): string | undefined {
	const keys = Object.keys(sortBy) as (keyof SortBy)[]
	const defaultKeys = Object.keys(defaultSort) as (keyof SortBy)[]
	const isDefault = keys.length === defaultKeys.length &&
		keys.every(k => sortBy[k] === defaultSort[k])
	if (isDefault) return undefined
	return keys.map(k => `${k}:${sortBy[k]}`).join(',')
}

const SORT_BY_DEFAULT: SortBy = {
	id: 'desc',
}

interface TaskListQueryState {
	sort: string | undefined
	filter: string | undefined
	s: string | undefined
	page: number
}

export function buildStoredQuery(state: TaskListQueryState): LocationQueryRaw {
	const query: LocationQueryRaw = {}
	if (state.sort) query.sort = state.sort
	if (state.filter) query.filter = state.filter
	if (state.s) query.s = state.s
	if (state.page > 1) query.page = String(state.page)
	return query
}

export function hasAnyTaskListQueryParam(query: LocationQuery): boolean {
	return query.sort !== undefined ||
		query.filter !== undefined ||
		query.s !== undefined ||
		query.page !== undefined
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

	const router = useRouter()
	const viewFiltersStore = useViewFiltersStore()

	const params = ref<TaskFilterParams>({...getDefaultTaskFilterParams()})

	// Restore stored query before the route-driven refs are read by the
	// load-tasks watcher below, so loadTasks fires once with the restored
	// values rather than twice (once empty, once restored).
	const initialQuery = router.currentRoute.value.query
	if (!hasAnyTaskListQueryParam(initialQuery)) {
		const storedQuery = viewFiltersStore.getViewQuery(projectViewIdGetter())
		if (Object.keys(storedQuery).length > 0) {
			router.replace({query: storedQuery})
		}
	}

	const page = useRouteQuery('page', '1', { transform: Number })
	const filter = useRouteQuery('filter')
	const s = useRouteQuery('s')

	watch(filter, v => { params.value.filter = v ?? '' }, { immediate: true })
	watch(s, v => { params.value.s = v ?? '' }, { immediate: true })

	watch(() => params.value.filter, v => { filter.value = v || undefined })
	watch(() => params.value.s, v => { s.value = v || undefined })

	const sortQuery = useRouteQuery('sort')

	const sortBy = computed<SortBy>({
		get() {
			const raw = sortQuery.value as string | undefined
			if (!raw) return {...sortByDefault}
			return parseSortQuery(raw, sortByDefault)
		},
		set(val: SortBy) {
			sortQuery.value = serializeSortBy(val, sortByDefault) || undefined
		},
	})

	// Mirror the URL query bits this composable owns into the store so
	// in-project tab switches and sidebar re-visits can restore them.
	watch(
		[sortQuery, filter, s, page],
		([sortValue, filterValue, sValue, pageValue]) => {
			const query = buildStoredQuery({
				sort: sortValue as string | undefined,
				filter: filterValue as string | undefined,
				s: sValue as string | undefined,
				page: pageValue,
			})
			if (Object.keys(query).length > 0) {
				viewFiltersStore.setViewQuery(projectViewId.value, query)
			} else {
				viewFiltersStore.clearViewQuery(projectViewId.value)
			}
		},
		{immediate: true},
	)

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
