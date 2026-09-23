import type {Bucket, ProjectView, Task, TaskReminder, User} from '@/client/generated'
import {REPEAT_TYPES, type IRepeatAfter} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {secondsToPeriod, periodToSeconds} from '@/helpers/time/period'
import {getDateWithTime} from '@/helpers/time/getDateWithTime'
import {cleanupItemText, PREFIXES, type ParsedTaskText, type PrefixMode} from '@/modules/quickAddMagic'

export function createTaskDraft(data: Partial<Task> = {}): Task {
	return {
		...data,
		id: data.id ?? 0,
		description: data.description ?? '',
		done: data.done ?? false,
		priority: data.priority ?? 0,
		labels: data.labels ?? [],
		assignees: data.assignees ?? [],
		reminders: data.reminders ?? [],
		attachments: data.attachments ?? [],
		buckets: data.buckets ?? [],
		related_tasks: data.related_tasks ?? {},
		reactions: data.reactions ?? {},
		comments: data.comments ?? [],
		project_id: data.project_id ?? 0,
		bucket_id: data.bucket_id ?? 0,
		repeat_after: data.repeat_after ?? 0,
		repeat_mode: data.repeat_mode ?? 0,
		percent_done: data.percent_done ?? 0,
		hex_color: data.hex_color ?? '',
		is_favorite: data.is_favorite ?? false,
		cover_image_attachment_id: data.cover_image_attachment_id ?? 0,
		title: (data.title ?? '').trim(),
	}
}

export function createReminderDraft(data: Partial<TaskReminder> = {}): TaskReminder {
	return {
		relative_period: data.relative_period ?? 0,
		relative_to: data.relative_to ?? '',
		reminder: data.reminder ?? '',
	}
}

export function getTaskIdentifier(task: Pick<Task, 'identifier' | 'index'> | null | undefined): string {
	if (!task) return ''
	return task.identifier && task.identifier !== `-${task.index}`
		? task.identifier
		: `#${task.index ?? 0}`
}

export function getHexColor(color?: string): string | undefined {
	if (!color || color === '#') return undefined
	return color.startsWith('#') ? color : `#${color}`
}

export function parseRepeatAfter(seconds = 0): IRepeatAfter {
	const {unit, amount} = secondsToPeriod(seconds)
	return {type: unit, amount}
}

// Nested related tasks are plain API objects, so repeat arrives as a parsed object, raw seconds, or not at all.
export function repeatAfterToSeconds(repeat?: number | IRepeatAfter | null): number {
	if (typeof repeat === 'number') return repeat
	return repeat?.amount ? periodToSeconds(repeat.amount, repeat.type) : 0
}

export function mergeTask(task: Task, updated: Task): Task {
	return {
		...task,
		...updated,
		// Write responses carry no view context, so cached per-view values win.
		position: task.position ?? updated.position,
		bucket_id: task.bucket_id ?? updated.bucket_id,
		assignees: updated.assignees ?? task.assignees,
		labels: updated.labels ?? task.labels,
		attachments: updated.attachments ?? task.attachments,
		related_tasks: updated.related_tasks ?? task.related_tasks,
		reactions: updated.reactions ?? task.reactions,
		created_by: updated.created_by ?? task.created_by,
	}
}

export function mapTasksDeep(tasks: readonly Task[], id: number, update: (task: Task) => Task): Task[] {
	return tasks.map(task => {
		// The updater's relations may contain the task itself again, so never descend into them.
		if (task.id === id) return update(task)
		if (!task.related_tasks) return task
		return {
			...task,
			related_tasks: Object.fromEntries(
				Object.entries(task.related_tasks)
					.map(([kind, children]) => [kind, mapTasksDeep(children ?? [], id, update)]),
			),
		}
	})
}

export function removeTask(tasks: readonly Task[], id: number): Task[] {
	return tasks
		.filter(task => task.id !== id)
		.map(task => task.related_tasks ? {
			...task,
			related_tasks: Object.fromEntries(
				Object.entries(task.related_tasks)
					.map(([kind, children]) => [kind, removeTask(children ?? [], id)]),
			),
		} : task)
}

export function getDefaultBucketId(
	view: Pick<ProjectView, 'default_bucket_id'>,
	buckets: readonly Bucket[],
): number | undefined {
	return view.default_bucket_id || buckets[0]?.id
}

export function moveTaskToBucket(buckets: Bucket[], task: Task, bucketId: number): Bucket[] {
	let match: {bucket: Bucket, card: Task} | undefined
	for (const bucket of buckets) {
		const card = bucket.tasks?.find(item => item.id === task.id)
		if (card) {
			match = {bucket, card}
			break
		}
	}
	if (!match || match.bucket.id === bucketId || !buckets.some(bucket => bucket.id === bucketId)) return buckets
	const {bucket: source, card} = match
	return buckets.map(bucket => {
		if (bucket.id === source.id) return {
			...bucket,
			count: Math.max(0, (bucket.count ?? 0) - 1),
			tasks: (bucket.tasks ?? []).filter(item => item.id !== task.id),
		}
		if (bucket.id === bucketId) return {
			...bucket,
			count: (bucket.count ?? 0) + 1,
			tasks: [{...mergeTask(card, task), bucket_id: bucketId}, ...(bucket.tasks ?? [])],
		}
		return bucket
	})
}

export function buildDefaultRemindersForQuickAdd(
	defaults: readonly TaskReminder[] | undefined,
	dueDate?: string | null,
): TaskReminder[] {
	return dueDate
		? (defaults ?? []).map(reminder => ({
			relative_period: reminder.relative_period,
			relative_to: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
		}))
		: []
}

export function buildQuickAddTask(
	parsed: ParsedTaskText,
	input: Partial<Task>,
	prefixes: (typeof PREFIXES)[PrefixMode] | undefined,
	assignees: (User & {match: string})[],
	defaults?: readonly TaskReminder[],
): Task {
	if (!parsed.text) return createTaskDraft(input)
	const title = prefixes?.assignee
		? cleanupItemText(parsed.text, assignees.map(user => user.match), prefixes.assignee)
		: parsed.text
	// Without a due date, a repeating task never becomes due.
	const dueDate = (parsed.date ?? (parsed.repeats ? getDateWithTime(new Date()) : null))?.toISOString()
	return createTaskDraft({
		...input,
		title,
		due_date: dueDate,
		priority: parsed.priority ?? 0,
		assignees: assignees.map(({match: _match, ...user}) => user),
		repeat_after: repeatAfterToSeconds(parsed.repeats),
		repeat_mode: parsed.repeats?.type === REPEAT_TYPES.Months && parsed.repeats.amount === 1
			? TASK_REPEAT_MODES.REPEAT_MODE_MONTH
			: TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT,
		reminders: buildDefaultRemindersForQuickAdd(defaults, dueDate),
	})
}
