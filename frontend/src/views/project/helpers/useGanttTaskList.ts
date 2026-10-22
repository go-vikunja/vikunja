import {computed, type Ref} from 'vue'
import type {Task} from '@/client/generated'
import type {Filters} from '@/composables/useRouteFilters'
import {useQuery} from '@tanstack/vue-query'
import {allTasksQuery} from '@/client/queries/tasks'
import type {TaskFilterParams} from '@/client/queries/tasks'
import {useCreateTaskMutation, useUpdateTaskMutation} from '@/client/queries/taskMutations'
import {useAuthStore} from '@/stores/auth'

export function useGanttTaskList<F extends Filters>(
	filters: Ref<F>,
	filterToApiParams: (filters: F) => TaskFilterParams,
	viewId: Ref<number>,
) {
	const auth = useAuthStore()
	const query = useQuery(computed(() => allTasksQuery({
		project: filters.value.projectId,
		view: viewId.value,
		params: {...filterToApiParams(filters.value), filter_timezone: auth.settings.timezone},
	})))
	const tasks = computed(() => new Map((query.data.value ?? []).map(task => [task.id, task])))
	const create = useCreateTaskMutation()
	const update = useUpdateTaskMutation(true)
	return {
		tasks, isLoading: query.isFetching,
		loadTasks: async () => { await query.refetch() },
		addTask: (task: Task) => create.mutateAsync({...task, project_id: filters.value.projectId}),
		updateTask: async (task: Task) => {
			const current = tasks.value.get(task.id!)
			if (current) await update.mutateAsync({...current, ...task, id: task.id!})
		},
	}
}

export type UseGanttTaskListReturn = ReturnType<typeof useGanttTaskList>
