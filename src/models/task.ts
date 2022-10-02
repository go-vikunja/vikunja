
import { PRIORITIES, type Priority } from '@/constants/priorities'

import type {ITask} from '@/modelTypes/ITask'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {IList} from '@/modelTypes/IList'
import type {ISubscription} from '@/modelTypes/ISubscription'
import type {IBucket} from '@/modelTypes/IBucket'

import type {IRepeatAfter} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES, type IRepeatMode} from '@/types/IRepeatMode'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import type { IRelationKind } from '@/types/IRelationKind'

import AbstractModel from './abstractModel'
import LabelModel from './label'
import UserModel from './user'
import AttachmentModel from './attachment'
import SubscriptionModel from './subscription'

export const TASK_DEFAULT_COLOR = '#1973ff'

const SUPPORTS_TRIGGERED_NOTIFICATION = 'Notification' in window && 'showTrigger' in Notification.prototype
if (!SUPPORTS_TRIGGERED_NOTIFICATION) {
	console.debug('This browser does not support triggered notifications')
}

export function	getHexColor(hexColor: string): string {
	if (hexColor === '' || hexColor === '#') {
		return TASK_DEFAULT_COLOR
	}

	return hexColor
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
	identifier = ''
	index = 0
	isFavorite = false
	subscription: ISubscription = null

	position = 0
	kanbanPosition = 0

	createdBy: IUser = UserModel
	created: Date = null
	updated: Date = null

	listId: IList['id'] = 0
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
		this.parseRepeatAfter()

		this.reminderDates = this.reminderDates.map(d => new Date(d))

		// Cancel all scheduled notifications for this task to be sure to only have available notifications
		this.cancelScheduledNotifications().then(() => {
			// Every time we see a reminder, we schedule a notification for it
			this.reminderDates.forEach(d => this.scheduleNotification(d))
		})

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

		// Set the task identifier to empty if the list does not have one
		if (this.identifier === `-${this.index}`) {
			this.identifier = ''
		}

		if (typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

		this.listId = Number(this.listId)
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

	/////////////////
	// Helper functions
	///////////////

	/**
	 * Parses the "repeat after x seconds" from the task into a usable js object inside the task.
	 * This function should only be called from the constructor.
	 */
	parseRepeatAfter() {
		const repeatAfterHours = (this.repeatAfter as number / 60) / 60
		this.repeatAfter = {type: 'hours', amount: repeatAfterHours}

		// if its dividable by 24, its something with days, otherwise hours
		if (repeatAfterHours % 24 === 0) {
			const repeatAfterDays = repeatAfterHours / 24
			if (repeatAfterDays % 7 === 0) {
				this.repeatAfter.type = 'weeks'
				this.repeatAfter.amount = repeatAfterDays / 7
			} else if (repeatAfterDays % 30 === 0) {
				this.repeatAfter.type = 'months'
				this.repeatAfter.amount = repeatAfterDays / 30
			} else if (repeatAfterDays % 365 === 0) {
				this.repeatAfter.type = 'years'
				this.repeatAfter.amount = repeatAfterDays / 365
			} else {
				this.repeatAfter.type = 'days'
				this.repeatAfter.amount = repeatAfterDays
			}
		}
	}

	async cancelScheduledNotifications() {
		if (!SUPPORTS_TRIGGERED_NOTIFICATION) {
			return
		}

		if (typeof navigator.serviceWorker === 'undefined') {
			console.debug('Service Worker not available')
			return
		}

		const registration = await navigator.serviceWorker.getRegistration()
		if (typeof registration === 'undefined') {
			return
		}

		// Get all scheduled notifications for this task and cancel them
		const scheduledNotifications = await registration.getNotifications({
			tag: `vikunja-task-${this.id}`,
			includeTriggered: true,
		})
		console.debug('Already scheduled notifications:', scheduledNotifications)
		scheduledNotifications.forEach(n => n.close())
	}

	async scheduleNotification(date) {
		if (typeof navigator.serviceWorker === 'undefined') {
			console.debug('Service Worker not available')
			return
		}

		if (date < new Date()) {
			console.debug('Date is in the past, not scheduling a notification. Date is ', date)
			return
		}

		if (!SUPPORTS_TRIGGERED_NOTIFICATION) {
			return
		}

		const {state} = await navigator.permissions.request({name: 'notifications'})
		if (state !== 'granted') {
			console.debug('Notification permission not granted, not showing notifications')
			return
		}

		const registration = await navigator.serviceWorker.getRegistration()
		if (typeof registration === 'undefined') {
			console.error('No service worker registration available')
			return
		}

		// Register the actual notification
		try {
			registration.showNotification('Vikunja Reminder', {
				tag: `vikunja-task-${this.id}`, // Group notifications by task id so we're only showing one notification per task
				body: this.title,
				// eslint-disable-next-line no-undef
				showTrigger: new TimestampTrigger(date),
				badge: '/images/icons/badge-monochrome.png',
				icon: '/images/icons/android-chrome-512x512.png',
				data: {taskId: this.id},
				actions: [
					{
						action: 'show-task',
						title: 'Show task',
					},
				],
			})
			console.debug('Notification scheduled for ' + date)
		} catch (e) {
			throw new Error('Error scheduling notification', e)
		}
	}
}

