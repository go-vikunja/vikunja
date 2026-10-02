import {allTasksQuery, type TaskFilterParams, type TaskResponse, type TaskScope} from '@/client/queries/tasks'
import {queryClient} from '@/client/queryClient'

export type SortDirection = 'asc' | 'desc'
export type SortMap = Record<string, SortDirection | undefined>

// Turns the sort of a view ({due_date: 'asc'}) into the two list parameters the API takes.
export function sortParams(sortBy: SortMap | undefined): Pick<TaskFilterParams, 'sort_by' | 'order_by'> {
	const keys = Object.keys(sortBy ?? {}).filter(key => sortBy?.[key])
	if (keys.length === 0) {
		return {}
	}
	return {
		sort_by: keys as TaskFilterParams['sort_by'],
		order_by: keys.map(key => sortBy?.[key] as SortDirection) as TaskFilterParams['order_by'],
	}
}

// The tasks of a view for an export: every page of what the view's filter and sort select, loaded
// fresh. The kanban view uses the whole project, its own endpoint answers with buckets.
export async function loadTasksForExport(
	projectId: number,
	viewId: number,
	kind: 'table' | 'list' | 'kanban' | 'gantt',
	params: TaskFilterParams,
	sortBy?: SortMap,
): Promise<TaskResponse[]> {
	const scope: TaskScope = {
		project: projectId,
		view: kind === 'kanban' ? 0 : viewId,
		params: {...params, ...sortParams(sortBy), expand: []},
	}
	return queryClient.fetchQuery({...allTasksQuery(scope), staleTime: 0})
}
