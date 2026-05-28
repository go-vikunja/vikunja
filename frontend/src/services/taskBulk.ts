import TaskService from '@/services/task'
import LabelTaskService from '@/services/labelTask'
import TaskAssigneeService from '@/services/taskAssignee'
import TaskDuplicateService from '@/services/taskDuplicateService'
import SubscriptionService from '@/services/subscription'
import TaskRelationService from '@/services/taskRelation'

import LabelTaskModel from '@/models/labelTask'
import TaskAssigneeModel from '@/models/taskAssignee'
import TaskDuplicateModel from '@/models/taskDuplicateModel'
import SubscriptionModel from '@/models/subscription'
import TaskRelationModel from '@/models/taskRelation'

import type {ITask} from '@/modelTypes/ITask'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import type {IProject} from '@/modelTypes/IProject'
import type {IRelationKind} from '@/types/IRelationKind'

function sleep(ms: number) {
	return new Promise(resolve => setTimeout(resolve, ms))
}

function isIgnorableDuplicateError(error: unknown): boolean {
	const maybeError = error as {
		response?: {
			status?: number,
			data?: {
				message?: string,
			},
		},
		message?: string,
	}

	const status = maybeError.response?.status
	const message = maybeError.response?.data?.message ?? maybeError.message ?? ''

	return status === 409 ||
		status === 412 ||
		message.toLowerCase().includes('already exists') ||
		message.toLowerCase().includes('duplicate')
}

function isDatabaseLockedError(error: unknown): boolean {
	const maybeError = error as {
		response?: {
			data?: {
				message?: string,
			},
		},
		message?: string,
	}

	const message = maybeError.response?.data?.message ?? maybeError.message ?? ''

	return message.toLowerCase().includes('database is locked')
}

export default class TaskBulkService {
	taskService = new TaskService()
	labelTaskService = new LabelTaskService()
	taskAssigneeService = new TaskAssigneeService()
	taskDuplicateService = new TaskDuplicateService()
	subscriptionService = new SubscriptionService()
	taskRelationService = new TaskRelationService()

	loading = false

	private async runWrite<T>(
		action: () => Promise<T>,
		options: {
			ignoreDuplicates?: boolean,
			retries?: number,
		} = {},
	): Promise<T | null> {
		const retries = options.retries ?? 5

		for (let attempt = 0; attempt <= retries; attempt++) {
			try {
				return await action()
			} catch (error) {
				if (options.ignoreDuplicates && isIgnorableDuplicateError(error)) {
					return null
				}

				if (isDatabaseLockedError(error) && attempt < retries) {
					await sleep(250 * (attempt + 1))
					continue
				}

				throw error
			}
		}

		return null
	}

	private async runSequential<T>(
		items: T[],
		action: (item: T) => Promise<unknown>,
		options: {
			ignoreDuplicates?: boolean,
			delayMs?: number,
		} = {},
	) {
		const delayMs = options.delayMs ?? 125

		for (const item of items) {
			await this.runWrite(
				() => action(item),
				{
					ignoreDuplicates: options.ignoreDuplicates,
				},
			)

			if (delayMs > 0) {
				await sleep(delayMs)
			}
		}
	}

	async updateTasks(tasks: ITask[], values: Partial<ITask>): Promise<ITask[]> {
		this.loading = true

		try {
			const updatedTasks: ITask[] = []

			await this.runSequential(tasks, async task => {
				const updatedTask = await this.taskService.update({
					...task,
					...values,
				}) as ITask

				updatedTasks.push(updatedTask)
			})

			return updatedTasks
		} finally {
			this.loading = false
		}
	}

	async setFavorite(tasks: ITask[], favorite: boolean): Promise<ITask[]> {
		return this.updateTasks(tasks, {
			isFavorite: favorite,
		})
	}

	async moveTasks(tasks: ITask[], project: IProject): Promise<ITask[]> {
		return this.updateTasks(tasks, {
			projectId: project.id,
		})
	}

	async deleteTasks(tasks: ITask[]) {
		this.loading = true

		try {
			await this.runSequential(
				tasks,
				task => this.taskService.delete(task),
				{
					delayMs: 175,
				},
			)
		} finally {
			this.loading = false
		}
	}

	async subscribe(tasks: ITask[]) {
		this.loading = true

		try {
			await this.runSequential(
				tasks.filter(task => task.subscription === null),
				task => this.subscriptionService.create(new SubscriptionModel({
					entity: 'task',
					entityId: task.id,
				})),
				{
					ignoreDuplicates: true,
				},
			)
		} finally {
			this.loading = false
		}
	}

	async unsubscribe(tasks: ITask[]) {
		this.loading = true

		try {
			await this.runSequential(
				tasks.filter(task => task.subscription !== null),
				task => this.subscriptionService.delete(new SubscriptionModel({
					entity: 'task',
					entityId: task.id,
				})),
				{
					ignoreDuplicates: true,
				},
			)
		} finally {
			this.loading = false
		}
	}

	async addAssignees(tasks: ITask[], users: IUser[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const usersToAdd = users.filter(user =>
					!(task.assignees ?? []).some(existing => existing.id === user.id),
				)

				await this.runSequential(
					usersToAdd,
					user => this.taskAssigneeService.create(new TaskAssigneeModel({
						taskId: task.id,
						userId: user.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async removeAssignees(tasks: ITask[], users: IUser[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const usersToRemove = users.filter(user =>
					(task.assignees ?? []).some(existing => existing.id === user.id),
				)

				await this.runSequential(
					usersToRemove,
					user => this.taskAssigneeService.delete(new TaskAssigneeModel({
						taskId: task.id,
						userId: user.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async replaceAssignees(tasks: ITask[], users: IUser[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const currentAssignees = task.assignees ?? []

				const usersToRemove = currentAssignees.filter(existing =>
					!users.some(user => user.id === existing.id),
				)

				const usersToAdd = users.filter(user =>
					!currentAssignees.some(existing => existing.id === user.id),
				)

				await this.runSequential(
					usersToRemove,
					user => this.taskAssigneeService.delete(new TaskAssigneeModel({
						taskId: task.id,
						userId: user.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)

				await this.runSequential(
					usersToAdd,
					user => this.taskAssigneeService.create(new TaskAssigneeModel({
						taskId: task.id,
						userId: user.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async addLabels(tasks: ITask[], labels: ILabel[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const labelsToAdd = labels.filter(label =>
					!(task.labels ?? []).some(existing => existing.id === label.id),
				)

				await this.runSequential(
					labelsToAdd,
					label => this.labelTaskService.create(new LabelTaskModel({
						taskId: task.id,
						labelId: label.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async removeLabels(tasks: ITask[], labels: ILabel[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const labelsToRemove = labels.filter(label =>
					(task.labels ?? []).some(existing => existing.id === label.id),
				)

				await this.runSequential(
					labelsToRemove,
					label => this.labelTaskService.delete(new LabelTaskModel({
						taskId: task.id,
						labelId: label.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async replaceLabels(tasks: ITask[], labels: ILabel[]) {
		this.loading = true

		try {
			for (const task of tasks) {
				const currentLabels = task.labels ?? []

				const labelsToRemove = currentLabels.filter(existing =>
					!labels.some(label => label.id === existing.id),
				)

				const labelsToAdd = labels.filter(label =>
					!currentLabels.some(existing => existing.id === label.id),
				)

				await this.runSequential(
					labelsToRemove,
					label => this.labelTaskService.delete(new LabelTaskModel({
						taskId: task.id,
						labelId: label.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)

				await this.runSequential(
					labelsToAdd,
					label => this.labelTaskService.create(new LabelTaskModel({
						taskId: task.id,
						labelId: label.id,
					})),
					{
						ignoreDuplicates: true,
					},
				)
			}
		} finally {
			this.loading = false
		}
	}

	async duplicate(tasks: ITask[]) {
		this.loading = true

		try {
			await this.runSequential(
				tasks,
				task => this.taskDuplicateService.create(new TaskDuplicateModel({
					taskId: task.id,
				})),
			)
		} finally {
			this.loading = false
		}
	}

	async addRelation(tasks: ITask[], otherTaskId: ITask['id'], relationKind: IRelationKind) {
		this.loading = true

		try {
			await this.runSequential(
				tasks.filter(task => task.id !== otherTaskId),
				task => this.taskRelationService.create(new TaskRelationModel({
					taskId: task.id,
					otherTaskId,
					relationKind,
				})),
				{
					ignoreDuplicates: true,
				},
			)
		} finally {
			this.loading = false
		}
	}
}