import {PRIORITIES, type Priority} from '@/constants/priorities'
import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_MONTH, SECONDS_A_WEEK, SECONDS_A_YEAR} from '@/constants/date'

import type {ITask} from '@/modelTypes/ITask'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {IProject} from '@/modelTypes/IProject'
import type {ISubscription} from '@/modelTypes/ISubscription'
import type {IBucket} from '@/modelTypes/IBucket'

import type {IRepeatAfter} from '@/types/IRepeatAfter'
import type {IRelationKind} from '@/types/IRelationKind'
import {TASK_REPEAT_MODES, type IRepeatMode} from '@/types/IRepeatMode'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'

import AbstractModel from './abstractModel'
import LabelModel from './label'
import UserModel from './user'
import AttachmentModel from './attachment'
import SubscriptionModel from './subscription'

export const TASK_DEFAULT_COLOR = '#1973ff'

export function	getHexColor(hexColor: string): string {
	if (hexColor === '' || hexColor === '#') {
		return TASK_DEFAULT_COLOR
	}

	return hexColor
}

/**
 * Parses `repeatAfterSeconds` into a usable js object.
 */
export function parseRepeatAfter(repeatAfterSeconds: number): IRepeatAfter {
	let repeatAfter: IRepeatAfter = {type: 'hours', amount: repeatAfterSeconds / SECONDS_A_HOUR}

	// if its dividable by 24, its something with days, otherwise hours
	if (repeatAfterSeconds % SECONDS_A_DAY === 0) {
		if (repeatAfterSeconds % SECONDS_A_WEEK === 0) {
			repeatAfter = {type: 'weeks', amount: repeatAfterSeconds / SECONDS_A_WEEK}
		} else if (repeatAfterSeconds % SECONDS_A_MONTH === 0) {
			repeatAfter = {type:'months', amount: repeatAfterSeconds / SECONDS_A_MONTH}
		} else if (repeatAfterSeconds % SECONDS_A_YEAR === 0) {
			repeatAfter = {type: 'years', amount: repeatAfterSeconds / SECONDS_A_YEAR}
		} else {
			repeatAfter = {type: 'days', amount: repeatAfterSeconds / SECONDS_A_DAY}
		}
	}
	return repeatAfter
}

export default class TaskModel extends AbstractModel<ITask> implements ITask {
	id = 0
	title = ''
	description = ''
	done = false
	doneAt: Date | null = null
	priority: Priority = PRIORITIES.UNSET
	labels: ILabel[] = []
	assignees: IUser[] = []

	dueDate: Date | null = 0
	startDate: Date | null = 0
	endDate: Date | null = 0
	repeatAfter: number | IRepeatAfter = 0
	repeatFromCurrentDate = false
	repeatMode: IRepeatMode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	reminderDates: Date[] = []
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
	coverImageAttachmentId: IAttachment['id'] = null

	position = 0
	kanbanPosition = 0

	createdBy: IUser = UserModel
	created: Date = null
	updated: Date = null

	projectId: IProject['id'] = 0
	bucketId: IBucket['id'] = 0

	constructor(data: Partial<ITask> = {}) {
		super()
		this.assignData(data)

		this.id = Number(this.id)
		this.title = this.title?.trim()
		this.doneAt = parseDateOrNull(this.doneAt)

		this.labels = this.labels
			.map(l => new LabelModel(l))
			.sort((f, s) => f.title > s.title ? 1 : -1)

		// Parse the assignees into user models
		this.assignees = this.assignees.map(a => {
			return new UserModel(a)
		})

		this.dueDate = parseDateOrNull(this.dueDate)
		this.startDate = parseDateOrNull(this.startDate)
		this.endDate = parseDateOrNull(this.endDate)

		// Parse the repeat after into something usable
		this.repeatAfter = parseRepeatAfter(this.repeatAfter as number)

		this.reminderDates = this.reminderDates.map(d => new Date(d))

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
	}

	getTextIdentifier() {
		if (this.identifier === '') {
			return `#${this.index}`
		}

		return this.identifier
	}

	getHexColor() {
		return getHexColor(this.hexColor)
	}
}

