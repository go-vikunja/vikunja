import {mutationOptions, queryOptions} from '@tanstack/vue-query'
import type {QueryClient} from '@tanstack/vue-query'

import {client} from '@/client/generated/client.gen'
import {tasksRelationsCreate, tasksRelationsDelete} from '@/client/generated'
import {taskKeys} from '@/client/queries/tasks'

// /api/v2/tasks/{task}/reschedule and /api/v2/projects/{project}/baseline are not in the generated
// SDK until `mage generate:frontend-client` was run (needs Go). Until then they go through the same
// generated runtime client, so authentication, token refresh and the identity fencing in
// client/http.ts apply. The wire types mirror pkg/routes/api/v2/task_schedule.go.

export interface RescheduledTask {
	id: number
	project_id: number
	start_date?: string
	end_date?: string
}

export interface RescheduleResult {
	task: RescheduledTask
	changed: RescheduledTask[]
	skipped: number[]
}

export interface BaselineEntry {
	task_id: number
	start_date?: string
	end_date?: string
}

export interface ProjectBaseline {
	saved_at?: string
	items: BaselineEntry[]
}

export const ganttKeys = {
	all: ['gantt'] as const,
	baseline: (projectId: number) => [...ganttKeys.all, 'baseline', projectId] as const,
}

async function request<T>(method: 'get' | 'post' | 'put' | 'delete', url: string, options: Record<string, unknown> = {}): Promise<T> {
	const result = await client[method]({url, ...options, throwOnError: true} as never) as {data: T}
	return result.data
}

export function projectBaselineQuery(projectId: () => number, enabled: () => boolean = () => true) {
	return queryOptions({
		queryKey: ganttKeys.baseline(projectId()),
		queryFn: () => request<ProjectBaseline>('get', `/projects/${projectId()}/baseline`),
		enabled: enabled() && projectId() > 0,
	})
}

// A move changes more than one task (the successors), so every task list is refreshed. Cache writes
// stay in the callbacks, with the client they are given.
function invalidateTasks(queryClient: QueryClient) {
	return queryClient.invalidateQueries({queryKey: taskKeys.all})
}

export interface RescheduleInput {
	id: number
	start_date: string | null
	end_date: string | null
}

export function rescheduleTaskMutation() {
	return mutationOptions({
		mutationFn: ({id, start_date, end_date}: RescheduleInput) => request<RescheduleResult>('post', `/tasks/${id}/reschedule`, {
			body: {start_date, end_date},
			headers: {'Content-Type': 'application/json'},
		}),
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateTasks(context.client),
	})
}

export function saveBaselineMutation() {
	return mutationOptions({
		mutationFn: (projectId: number) => request<{tasks: number}>('put', `/projects/${projectId}/baseline`),
		onSettled: (_data, _error, projectId, _ctx, context) => context.client.invalidateQueries({queryKey: ganttKeys.baseline(projectId)}),
	})
}

export function clearBaselineMutation() {
	return mutationOptions({
		mutationFn: (projectId: number) => request<void>('delete', `/projects/${projectId}/baseline`),
		onSettled: (_data, _error, projectId, _ctx, context) => context.client.invalidateQueries({queryKey: ganttKeys.baseline(projectId)}),
	})
}

export interface DependencyInput {
	// The predecessor: it has to end before `successorId` starts.
	predecessorId: number
	successorId: number
}

export function createDependencyMutation() {
	return mutationOptions({
		mutationFn: async ({predecessorId, successorId}: DependencyInput) => (await tasksRelationsCreate({
			path: {task: predecessorId},
			body: {other_task_id: successorId, relation_kind: 'precedes'},
		})).data,
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateTasks(context.client),
	})
}

export function deleteDependencyMutation() {
	return mutationOptions({
		mutationFn: async ({predecessorId, successorId}: DependencyInput) => (await tasksRelationsDelete({
			path: {task: predecessorId, relationKind: 'precedes', otherTask: successorId},
		})).data,
		onSettled: (_data, _error, _vars, _ctx, context) => invalidateTasks(context.client),
	})
}
