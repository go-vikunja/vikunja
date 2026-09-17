import {PRIORITIES, type Priority} from '@/constants/priorities'

import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ISubscription} from '@/modelTypes/ISubscription'
import type {IBucket} from '@/modelTypes/IBucket'

import type {IRepeatAfter} from '@/types/IRepeatAfter'
import type {IRelationKind} from '@/types/IRelationKind'
import {TASK_REPEAT_MODES, type IRepeatMode} from '@/types/IRepeatMode'
import type {Label, User} from '@/client/generated'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {objectToSnakeCase} from '@/helpers/case'

import AbstractModel from './abstractModel'
import UserModel from './user'
import AttachmentModel from './attachment'
import SubscriptionModel from './subscription'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import TaskReminderModel from '@/models/taskReminder'
import TaskCommentModel from '@/models/taskComment.ts'

import {getHexColor, getTaskIdentifier, parseRepeatAfter} from '@/helpers/task'
export {getHexColor, getTaskIdentifier, parseRepeatAfter} from '@/helpers/task'

export default class TaskModel extends AbstractModel<ITask> implements ITask {
	id = 0
	title = ''
	description = ''
	done = false
	doneAt: Date | null = null
	deletedAt: Date | null = null
	priority: Priority = PRIORITIES.UNSET
	labels: Label[] = []
	assignees: User[] = []

	dueDate: Date | null = 0
	startDate: Date | null = 0
	endDate: Date | null = 0
	repeatAfter: number | IRepeatAfter = 0
	repeatFromCurrentDate = false
	repeatMode: IRepeatMode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	reminders: ITaskReminder[] = []
	parentTaskId: ITask['id'] = 0
	hexColor = ''
	percentDone = 0
	relatedTasks:  Partial<Record<IRelationKind, ITask[]>> = {}
	attachments: IAttachment[] = []
	coverImageAttachmentId: IAttachment['id'] = null
	identifier = ''
	index = 0
	isFavorite = false
	subscription: ISubscription = null

	position = 0
	
	reactions = {}
	comments = []

	createdBy: IUser = UserModel
	created: Date = null
	updated: Date = null

	projectId = 0
	bucketId: IBucket['id'] = 0
	buckets: IBucket[] = []

	constructor(data: Partial<ITask> = {}) {
		super()
		const labels = (data.labels ?? []).map(label => objectToSnakeCase(label) as Label)
		const assignees = (data.assignees ?? []).map(user => objectToSnakeCase(user) as User)
		this.assignData(data)

		this.id = Number(this.id)
		this.title = this.title?.trim()
		this.doneAt = parseDateOrNull(this.doneAt)
		this.deletedAt = parseDateOrNull(this.deletedAt)

		this.labels = labels.sort((a, b) => (a.title ?? '').localeCompare(b.title ?? ''))

		this.assignees = assignees

		this.dueDate = parseDateOrNull(this.dueDate)
		this.startDate = parseDateOrNull(this.startDate)
		this.endDate = parseDateOrNull(this.endDate)

		// Parse the repeat after into something usable
		this.repeatAfter = parseRepeatAfter(this.repeatAfter as number)

		this.reminders = this.reminders.map(r => new TaskReminderModel(r))

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		// Convert all subtasks to task models
		Object.keys(this.relatedTasks).forEach(relationKind => {
			this.relatedTasks[relationKind] = this.relatedTasks[relationKind].map(t => {
				return new TaskModel(t)
			})
		})

		// Make all attachments to attachment models
		this.attachments = this.attachments.map(a => new AttachmentModel(a))

		// Set the task identifier to empty if the project does not have one
		if (this.identifier === `-${this.index}`) {
			this.identifier = ''
		}

		if (typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

		this.projectId = Number(this.projectId)

		// If we would use the camel cased value here, it would lose the reactions - emojis can't be camel cased.
		// The comments will be camel cased anyway in the constructor of the task comment model.
		this.comments = (data.comments || []).map(c => new TaskCommentModel(c))

		// We can't convert emojis to camel case, hence we do this manually
		this.reactions = {}
		Object.keys(data.reactions || {}).forEach(reaction => {
			this.reactions[reaction] = data.reactions[reaction].map(u => new UserModel(u))
		})
	}

	getTextIdentifier() {
		return getTaskIdentifier(this)
	}

	getHexColor() {
		return getHexColor(this.hexColor)
	}
}
