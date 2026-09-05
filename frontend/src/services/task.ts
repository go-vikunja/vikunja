import AbstractService from './abstractService'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import AttachmentService from './attachment'

import {colorFromHex} from '@/helpers/color/colorFromHex'
import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_WEEK} from '@/constants/date'
import {objectToSnakeCase} from '@/helpers/case'
import {apiV2Url, AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {invalidateCachedTask} from '@/helpers/taskCache'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'
import {translatedError} from '@/message'

// Mirrors models.MaxTasksPerBulkCreation on the backend.
const MAX_TASKS_PER_BULK_CREATION = 100

/**
 * Tasks reaching processModel did not necessarily go through the TaskModel
 * constructor - related tasks nested in a task are plain api objects - so
 * repeatAfter is either the parsed object, raw seconds, or missing entirely.
 */
function repeatAfterToSeconds(repeatAfter: ITask['repeatAfter'] | undefined): number {
	if (typeof repeatAfter === 'number') {
		return repeatAfter
	}

	if (!repeatAfter?.amount) {
		return 0
	}

	switch (repeatAfter.type) {
		case 'hours':
			return repeatAfter.amount * SECONDS_A_HOUR
		case 'days':
			return repeatAfter.amount * SECONDS_A_DAY
		case 'weeks':
			return repeatAfter.amount * SECONDS_A_WEEK
		default:
			return 0
	}
}

export default class TaskService extends AbstractService<ITask> {
	constructor() {
		super({
			create: '/projects/{projectId}/tasks',
			getAll: '/tasks',
			get: '/tasks/{id}',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		})
	}

	modelFactory(data) {
		return new TaskModel(data)
	}

	beforeUpdate(model) {
		return this.processModel(model)
	}

	beforeCreate(model) {
		return this.processModel(model)
	}

	autoTransformBeforePost(): boolean {
		return false
	}

	async update(model: ITask) {
		const updated = await super.update(model)
		invalidateCachedTask(model.id)
		return updated
	}

	async delete(model: ITask) {
		const response = await super.delete(model)
		invalidateCachedTask(model.id)
		return response
	}

	processModel(updatedModel) {
		const model = {...updatedModel}

		model.title = model.title?.trim()

		// Ensure that projectId is an int
		model.projectId = Number(model.projectId)

		// Convert dates into an iso string
		model.dueDate = toISOStringOrNull(model.dueDate)
		model.startDate = toISOStringOrNull(model.startDate)
		model.endDate = toISOStringOrNull(model.endDate)
		model.doneAt = toISOStringOrNull(model.doneAt)
		model.deletedAt = toISOStringOrNull(model.deletedAt)
		model.created = toISOStringOrNull(model.created)
		model.updated = toISOStringOrNull(model.updated)

		model.reminderDates = null
		// remove all nulls, these would create empty reminders
		model.reminders = (model.reminders ?? []).filter(r => r !== null)
		// Make normal timestamps from js dates
		if (model.reminders.length > 0) {
			model.reminders.forEach(r => {
				r.reminder = toISOStringOrNull(r.reminder)
			})
		}

		model.repeatAfter = repeatAfterToSeconds(model.repeatAfter)

		model.hexColor = colorFromHex(model.hexColor ?? '')

		// Do the same for all related tasks. `model` is only a shallow copy, so this
		// has to build a new object - assigning into relatedTasks would replace the
		// related tasks of the task we were passed with their api representation.
		model.relatedTasks = Object.fromEntries(
			Object.entries<ITask[]>(model.relatedTasks ?? {})
				.map(([relationKind, tasks]) => [relationKind, tasks.map(t => this.processModel(t))]),
		)

		// Process all attachments to prevent parsing errors
		if (model.attachments?.length > 0) {
			const attachmentService = new AttachmentService()
			model.attachments.map(a => {
				return attachmentService.processModel(a)
			})
		}

		const transformed = objectToSnakeCase(model)

		// We can't convert emojis to skane case, hence we add them back again
		transformed.reactions = {}
		Object.keys(updatedModel.reactions || {}).forEach(reaction => {
			transformed.reactions[reaction] = updatedModel.reactions[reaction].map(u => objectToSnakeCase(u))
		})

		return transformed as ITask
	}

	// The v2 endpoint validates strictly against the task schema and rejects the
	// frontend-only properties (max_permission, reminder_dates, …) processModel
	// adds, hence the allowlist.
	private toBulkCreatePayload(task: ITask) {
		// processModel lies about its return type — it returns the snake_cased
		// wire format, not an ITask.
		const processed = this.processModel(task) as unknown as {
			assignees: {id: number, username: string}[],
			reminders: {reminder: string | null, relative_period: number, relative_to: string | null}[],
		} & Record<string, unknown>
		return {
			title: processed.title,
			description: processed.description,
			done: processed.done,
			due_date: processed.due_date,
			start_date: processed.start_date,
			end_date: processed.end_date,
			priority: processed.priority,
			hex_color: processed.hex_color,
			percent_done: processed.percent_done,
			repeat_after: processed.repeat_after,
			repeat_mode: processed.repeat_mode,
			is_favorite: processed.is_favorite,
			bucket_id: processed.bucket_id,
			assignees: processed.assignees.map(a => ({
				id: a.id,
				username: a.username,
			})),
			reminders: processed.reminders.map(r => ({
				reminder: r.reminder,
				relative_period: r.relative_period,
				relative_to: r.relative_to,
			})),
		}
	}

	// Returns tasks aligned 1:1 with the input (null = not created). Grouped per
	// project because the endpoint takes the project from the URL.
	async bulkCreate(tasks: ITask[]): Promise<{tasks: (ITask | null)[], error: unknown | null}> {
		const cancel = this.setLoading()

		try {
			const groups = new Map<ITask['projectId'], number[]>()
			tasks.forEach((task, index) => {
				const group = groups.get(task.projectId)
				if (group) {
					group.push(index)
				} else {
					groups.set(task.projectId, [index])
				}
			})

			const created: (ITask | null)[] = new Array(tasks.length).fill(null)
			let error: unknown | null = null
			// Sequential throughout: the server assigns task indexes at insert time,
			// and concurrent bulk writes fail under write contention (SQLite).
			for (const [projectId, indexes] of groups) {
				const batches: number[][] = []
				for (let i = 0; i < indexes.length; i += MAX_TASKS_PER_BULK_CREATION) {
					batches.push(indexes.slice(i, i + MAX_TASKS_PER_BULK_CREATION))
				}

				// Last chunk first: the server puts each batch on top in every view, so
				// posting in reverse leaves the earliest input lines topmost. Tradeoff:
				// per-project index numbers then run backwards across batches.
				for (const batch of batches.reverse()) {
					try {
						// Fresh http instance: the shared one's interceptors would run
						// processModel on the {tasks} wrapper.
						const {data} = await AuthenticatedHTTPFactory().post(
							apiV2Url(`projects/${Number(projectId)}/tasks/bulk`),
							{tasks: batch.map(index => this.toBulkCreatePayload(tasks[index]))},
						)
						if (!Array.isArray(data?.tasks) || data.tasks.length !== batch.length) {
							throw translatedError('task.bulkCreateUnexpectedResponse')
						}
						// The response is in payload order. Don't match by title — quick
						// add magic cleans titles and duplicates would collide.
						data.tasks.forEach((t: Partial<ITask>, batchIndex: number) => {
							created[batch[batchIndex]] = this.modelCreateFactory(t)
						})
					} catch (e) {
						// Keep what other batches created so the caller can retry only
						// the missing tasks instead of duplicating everything.
						error ??= e
						break
					}
				}
			}

			return {tasks: created, error}
		} finally {
			cancel()
		}
	}

	async markTaskAsRead(taskId: ITask['id']): Promise<void> {
		const cancel = this.setLoading()
	
		try {
			await AuthenticatedHTTPFactory().post(`/tasks/${taskId}/read`, {} as ITask)
		} finally {
			cancel()
		}
	}
}
