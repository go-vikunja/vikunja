import {useMutation, type QueryClient} from '@tanstack/vue-query'
import {
	tasksCreate,
	patchTasksRead,
	tasksDelete,
	tasksBulkCreate,
	tasksDuplicate,
	tasksMarkRead,
	taskAssigneesCreate,
	taskAssigneesDelete,
	taskLabelsCreate,
	taskLabelsDelete,
	tasksRelationsCreate,
	tasksRelationsDelete,
	tasksPositionUpdate,
	taskBucketUpdate,
	type Task,
	type TaskWritable,
	type User,
	type Label,
	type TaskRelationWritable,
	type TasksRelationsDeleteData,
	type TaskPositionWritable,
	type JsonPatchOp,
	type Bucket,
} from '@/client/generated'
import {assertClientRequestContext, captureClientRequestContext} from '@/client/requestContext'
import {contextMutationOptions} from './contextMutation'
import {normalizeTask, taskKeys, type PaginatedTaskResponse, type TaskResponse} from './tasks'
import {bucketKeys, kanbanKeys, normalizeBucket, type BoardData} from './kanban'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
	removeTaskEverywhere,
	replaceTaskEverywhere,
	taskQueryKeys,
} from './taskCache'
import {getCachedProject, projectKeys} from './projects'
import {colorFromHex} from '@/helpers/color/colorFromHex'
import {getDefaultBucketId, moveTaskToBucket} from '@/helpers/task'
import {translatedError} from '@/message'

const API_ZERO_DATE = '0001-01-01T00:00:00Z'

const writableFields = [
	'title',
	'description',
	'done',
	'due_date',
	'start_date',
	'end_date',
	'priority',
	'hex_color',
	'percent_done',
	'repeat_after',
	'repeat_mode',
	'is_favorite',
	'bucket_id',
	'project_id',
	'reminders',
	'cover_image_attachment_id',
] as const satisfies readonly (keyof TaskWritable)[]

export function taskWriteBody(task: Task): TaskWritable {
	const body: TaskWritable = Object.fromEntries(
		writableFields
			.filter(field => task[field] !== undefined)
			.map(field => [field, task[field]]),
	)
	if (body.title !== undefined) body.title = body.title.trim()
	if (body.hex_color !== undefined) body.hex_color = colorFromHex(body.hex_color)
	for (const field of ['due_date', 'start_date', 'end_date'] as const) {
		if (body[field] === '') body[field] = API_ZERO_DATE
	}
	if (body.reminders) body.reminders = body.reminders.map(reminder => ({
		...reminder,
		reminder: reminder.reminder === '' ? API_ZERO_DATE : reminder.reminder,
	}))
	return body
}

function taskPatchBody(task: Task): JsonPatchOp[] {
	return Object.entries(taskWriteBody(task)).map(([field, value]) => ({op: 'add', path: `/${field}`, value}))
}

function reconcileDoneBuckets(client: QueryClient, task: TaskResponse) {
	const handled = new Set<number>()
	for (const [key, board] of client.getQueriesData<BoardData>({queryKey: kanbanKeys.all})) {
		const project = kanbanKeys.projectOf(key)
		const viewId = kanbanKeys.viewOf(key)
		if (!board || project === undefined || viewId === undefined) continue
		const view = getCachedProject(project, client)?.views?.find(item => item.id === viewId)
		if (!view?.done_bucket_id) continue
		const source = board.buckets.find(bucket => bucket.tasks.some(item => item.id === task.id))
		if (!source) continue
		handled.add(viewId)
		const target = task.done
			? view.done_bucket_id
			: source.id === view.done_bucket_id
				? getDefaultBucketId(view, board.buckets)
				: source.id
		if (target === undefined) continue
		const buckets = moveTaskToBucket(board.buckets, task, target)
		if (buckets === board.buckets) continue
		client.setQueryData(key, {...board, buckets: buckets.map(normalizeBucket)})
		reconcileTaskBuckets(client, {project, view: viewId, bucket: target, task: task.id})
	}
	reconcileDoneBucketsWithoutBoard(client, task, handled)
}

// Without a board holding the card, the detail's own buckets array is the only record of where it sits.
function reconcileDoneBucketsWithoutBoard(client: QueryClient, task: TaskResponse, handled: Set<number>) {
	for (const view of getCachedProject(task.project_id, client)?.views ?? []) {
		if (view.id === undefined || handled.has(view.id) || !view.done_bucket_id) continue
		const source = cachedTaskBucketId(client, task.id, view.id)
		if (source === undefined) continue
		const target = task.done
			? view.done_bucket_id
			: source === view.done_bucket_id
				? getDefaultBucketId(view, cachedViewBuckets(client, task.project_id, view.id) ?? [])
				: undefined
		if (target === undefined || target === source) continue
		reconcileTaskBuckets(client, {
			project: task.project_id,
			view: view.id,
			bucket: target,
			task: task.id,
		})
	}
}

function cachedTaskBucketId(client: QueryClient, task: number, view: number): number | undefined {
	for (const [, detail] of client.getQueriesData<TaskResponse>({queryKey: taskKeys.details})) {
		if (detail?.id !== task) continue
		const bucket = detail.buckets?.find(item => item.project_view_id === view)
		if (bucket?.id !== undefined) return bucket.id
	}
	return undefined
}

function cachedViewBuckets(client: QueryClient, project: number, view: number) {
	const board = client.getQueriesData<BoardData>({queryKey: kanbanKeys.view(project, view)})
		.find(([, data]) => data)?.[1]
	return board?.buckets ?? client.getQueryData<Bucket[]>(bucketKeys.list(project, view))
}

function insertTaskIntoBoards(client: QueryClient, created: Task) {
	const task = normalizeTask(created)
	for (const [key, board] of client.getQueriesData<BoardData>({queryKey: kanbanKeys.all})) {
		const project = kanbanKeys.projectOf(key)
		const viewId = kanbanKeys.viewOf(key)
		if (!board || project === undefined || project !== task.project_id || viewId === undefined) continue
		const view = getCachedProject(project, client)?.views?.find(item => item.id === viewId)
		// The response only echoes the requested bucket, which belongs to the one view it was requested for.
		const target = board.buckets.some(bucket => bucket.id === task.bucket_id)
			? task.bucket_id
			: view && view.bucket_configuration_mode !== 'filter'
				? getDefaultBucketId(view, board.buckets)
				: undefined
		if (target === undefined) continue
		client.setQueryData(key, {
			...board,
			buckets: board.buckets.map(bucket => bucket.id === target
				? {...bucket, count: bucket.count + 1, tasks: [{...task, bucket_id: target}, ...bucket.tasks]}
				: bucket),
		})
	}
}

export function createTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (task: Task & Required<Pick<Task, 'project_id'>>) => (await tasksCreate({
			path: {project: task.project_id},
			body: taskWriteBody(task),
		})).data,
		onSuccess: (task, _input, client) => insertTaskIntoBoards(client, task),
		onSettled: (_task, client) => invalidateTaskMembership(client, undefined, 'active'),
	})
}

export function updateTaskMutationOptions(optimistic = false) {
	return contextMutationOptions({
		mutationFn: async (task: Task & Required<Pick<Task, 'id'>>) => (await patchTasksRead({
			path: {task: task.id},
			body: taskPatchBody(task),
		})).data,
		optimistic: optimistic ? {
			queryKeys: ({id}, client) => taskQueryKeys(client, id),
			update: (task, client) => replaceTaskEverywhere(client, task),
		} : undefined,
		onSuccess: (task, input, client) => {
			replaceTaskEverywhere(client, task)
			reconcileDoneBuckets(client, normalizeTask({...task, id: input.id}))
		},
		onSettled: ({id}, client) => invalidateTaskMembership(client, id),
	})
}

export function deleteTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await tasksDelete({path: {task: id}}) },
		onSuccess: (_data, id, client) => removeTaskEverywhere(client, id),
		onSettled: (_id, client) => invalidateTaskMembership(client, undefined, 'active'),
	})
}

export function bulkCreateTasksMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (input: Task[]) => {
			const context = captureClientRequestContext()
			const tasks: (Task | null)[] = Array(input.length).fill(null)
			let error: unknown = null
			const groups = new Map<number, {task: Task, index: number}[]>()
			input.forEach((task, index) => {
				const project = task.project_id ?? 0
				groups.set(project, [...(groups.get(project) ?? []), {task, index}])
			})
			for (const [project, entries] of groups) {
				const batches = []
				for (let index = 0; index < entries.length; index += 100) batches.push(entries.slice(index, index + 100))
				// The API inserts batches at the top, so submit the last batch first.
				for (const batch of batches.reverse()) {
					assertClientRequestContext(context)
					try {
						const {data} = await tasksBulkCreate({
							path: {project},
							body: {tasks: batch.map(({task}) => taskWriteBody(task))},
						})
						if (data.tasks?.length !== batch.length) throw translatedError('task.bulkCreateUnexpectedResponse')
						data.tasks!.forEach((task, index) => { tasks[batch[index].index] = task })
					} catch (cause) {
						error ??= cause
						break
					}
				}
			}
			return {tasks, error}
		},
		// The backend keeps the input order top to bottom, so insert back to front.
		onSuccess: ({tasks}, _input, client) => [...tasks].reverse().forEach(task => {
			if (task) insertTaskIntoBoards(client, task)
		}),
		onSettled: (_input, client) => invalidateTaskMembership(client, undefined, 'active'),
	})
}

export function duplicateTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await tasksDuplicate({path: {task: id}})).data.duplicated_task!,
		onSuccess: (task, _id, client) => insertTaskIntoBoards(client, task),
		onSettled: (_id, client) => invalidateTaskMembership(client, undefined, 'active'),
	})
}

export function favoriteTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (task: Task & Required<Pick<Task, 'id'>>) => (await patchTasksRead({
			path: {task: task.id},
			body: taskPatchBody({is_favorite: !task.is_favorite}),
		})).data,
		optimistic: {
			queryKeys: ({id}, client) => taskQueryKeys(client, id),
			update: (task, client) => mapTaskEverywhere(client, task.id, current => ({
				...current,
				is_favorite: !task.is_favorite,
			})),
		},
		onSuccess: (task, _input, client) => replaceTaskEverywhere(client, task),
		onSettled: async ({id}, client) => {
			await invalidateTaskMembership(client, id)
			await client.invalidateQueries({queryKey: projectKeys.list()})
		},
	})
}

export function markTaskReadMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await tasksMarkRead({path: {task: id}}) },
		onSuccess: (_data, id, client) => mapTaskEverywhere(client, id, task => ({...task, is_unread: false})),
		onSettled: (id, client) => invalidateTaskMembership(client, id),
	})
}

type AssigneeInput = {taskId: number, user: User & Required<Pick<User, 'id'>>}
export function addTaskAssigneeMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, user}: AssigneeInput) =>
			(await taskAssigneesCreate({path: {task: taskId}, body: {user_id: user.id}})).data,
		onSuccess: (_data, {taskId, user}, client) => mapTaskEverywhere(client, taskId, task => ({
			...task,
			assignees: [...task.assignees.filter(item => item.id !== user.id), user],
		})),
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
	})
}
export function removeTaskAssigneeMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, user}: AssigneeInput) => {
			await taskAssigneesDelete({path: {task: taskId, user: user.id}})
		},
		onSuccess: (_data, {taskId, user}, client) => mapTaskEverywhere(client, taskId, task => ({
			...task,
			assignees: task.assignees.filter(item => item.id !== user.id),
		})),
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
	})
}

type LabelInput = {taskId: number, label: Label & Required<Pick<Label, 'id'>>}
export function addTaskLabelMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, label}: LabelInput) =>
			(await taskLabelsCreate({path: {task: taskId}, body: {label_id: label.id}})).data,
		onSuccess: (_data, {taskId, label}, client) => mapTaskEverywhere(client, taskId, task => ({
			...task,
			labels: [...task.labels.filter(item => item.id !== label.id), label],
		})),
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
	})
}
export function removeTaskLabelMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, label}: LabelInput) => {
			await taskLabelsDelete({path: {task: taskId, label: label.id}})
		},
		onSuccess: (_data, {taskId, label}, client) => mapTaskEverywhere(client, taskId, task => ({
			...task,
			labels: task.labels.filter(item => item.id !== label.id),
		})),
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
	})
}

export function createTaskRelationMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, ...body}: TaskRelationWritable & {taskId: number}) =>
			(await tasksRelationsCreate({path: {task: taskId}, body})).data,
		onSettled: async ({taskId, other_task_id}, client) => {
			await invalidateTaskMembership(client, taskId, 'active')
			if (other_task_id) await invalidateTaskMembership(client, other_task_id, 'active')
		},
	})
}
export function deleteTaskRelationMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (path: TasksRelationsDeleteData['path']) => {
			await tasksRelationsDelete({path})
		},
		onSettled: async ({task, otherTask}, client) => {
			await invalidateTaskMembership(client, task, 'active')
			await invalidateTaskMembership(client, otherTask, 'active')
		},
	})
}

export function updateTaskPositionMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({taskId, ...body}: TaskPositionWritable & {taskId: number}) =>
			(await tasksPositionUpdate({path: {task: taskId}, body})).data,
		onSuccess: (data, {taskId, project_view_id}, client) => {
			for (const [key, list] of client.getQueriesData<PaginatedTaskResponse>({queryKey: taskKeys.lists})) {
				const params = taskKeys.paramsOf(key)
				if (!list || !params || taskKeys.viewOf(key) !== project_view_id) continue
				const direction = params.order_by?.[0] === 'desc' ? -1 : 1
				const items = list.items.map(task => task.id === taskId
					? {...task, position: data.position ?? task.position}
					: task)
				client.setQueryData(key, {
					...list,
					items: params.sort_by?.[0] === 'position'
						? items.sort((a, b) => direction * (a.position - b.position))
						: items,
				})
			}
			for (const [key, board] of client.getQueriesData<BoardData>({queryKey: kanbanKeys.all})) {
				if (board && kanbanKeys.viewOf(key) === project_view_id) client.setQueryData(key, {
					...board,
					buckets: board.buckets.map(bucket => ({
						...bucket,
						tasks: bucket.tasks
							.map(task => task.id === taskId
								? {...task, position: data.position ?? task.position}
								: task)
							.sort((a, b) => a.position - b.position),
					})),
				})
			}
		},
		onSettled: ({taskId}, client) => invalidateTaskMembership(client, taskId),
	})
}

// The detail's buckets array is the only source of the bucket label and nothing refetches it.
function reconcileTaskBuckets(
	client: QueryClient,
	{project, view, bucket, task}: {project: number, view: number, bucket: number, task: number},
) {
	const target = cachedViewBuckets(client, project, view)?.find(item => item.id === bucket)
	if (!target) {
		void client.invalidateQueries({queryKey: [...taskKeys.details, task], refetchType: 'active'})
		return
	}
	const {tasks: _tasks, ...entry} = target
	mapTaskEverywhere(client, task, current => current.buckets
		? {
			...current,
			buckets: current.buckets.map(item => item.project_view_id === view
				? {...entry, project_view_id: view}
				: item),
		}
		: current)
}

export function moveTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({project, view, bucket, task}: {
			project: number,
			view: number,
			bucket: number,
			task: Task,
		}) => (await taskBucketUpdate({path: {project, view, bucket}, body: {task_id: task.id}})).data,
		optimistic: {
			queryKeys: ({project, view}) => [kanbanKeys.view(project, view)],
			update: ({project, view, bucket, task}, client) => client.setQueriesData<BoardData>(
				{queryKey: kanbanKeys.view(project, view)},
				current => current ? {
					...current,
					buckets: moveTaskToBucket(current.buckets, task, bucket).map(normalizeBucket),
				} : current,
			),
		},
		onSuccess: (data, {project, view, bucket, task}, client) => {
			const target = data.bucket_id ?? bucket
			if (data.task) replaceTaskEverywhere(client, data.task)
			client.setQueriesData<BoardData>(
				{queryKey: kanbanKeys.view(project, view)},
				current => current ? {
					...current,
					buckets: moveTaskToBucket(current.buckets, data.task ?? task, target).map(normalizeBucket),
				} : current,
			)
			const id = data.task?.id ?? task.id
			if (id !== undefined) reconcileTaskBuckets(client, {project, view, bucket: target, task: id})
		},
		onSettled: ({task}, client) => invalidateTaskMembership(client, task.id),
	})
}

export function useCreateTaskMutation() {
	return useMutation(createTaskMutationOptions())
}

export function useUpdateTaskMutation(optimistic = false) {
	return useMutation(updateTaskMutationOptions(optimistic))
}

export function useDeleteTaskMutation() {
	return useMutation(deleteTaskMutationOptions())
}

export function useBulkCreateTasksMutation() {
	return useMutation(bulkCreateTasksMutationOptions())
}

export function useDuplicateTaskMutation() {
	return useMutation(duplicateTaskMutationOptions())
}

export function useFavoriteTaskMutation() {
	return useMutation(favoriteTaskMutationOptions())
}

export function useMarkTaskReadMutation() {
	return useMutation(markTaskReadMutationOptions())
}

export function useAddTaskAssigneeMutation() {
	return useMutation(addTaskAssigneeMutationOptions())
}

export function useRemoveTaskAssigneeMutation() {
	return useMutation(removeTaskAssigneeMutationOptions())
}

export function useAddTaskLabelMutation() {
	return useMutation(addTaskLabelMutationOptions())
}

export function useRemoveTaskLabelMutation() {
	return useMutation(removeTaskLabelMutationOptions())
}

export function useCreateTaskRelationMutation() {
	return useMutation(createTaskRelationMutationOptions())
}

export function useDeleteTaskRelationMutation() {
	return useMutation(deleteTaskRelationMutationOptions())
}

export function useUpdateTaskPositionMutation() {
	return useMutation(updateTaskPositionMutationOptions())
}

export function useMoveTaskMutation() {
	return useMutation(moveTaskMutationOptions())
}
