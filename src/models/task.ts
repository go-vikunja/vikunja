import AbstractModel from './abstractModel'
import UserModel from './user'
import LabelModel from './label'
import AttachmentModel from './attachment'
import {REPEAT_MODE_DEFAULT} from './constants/taskRepeatModes'

import SubscriptionModel from '@/models/subscription'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

const SUPPORTS_TRIGGERED_NOTIFICATION = 'Notification' in window && 'showTrigger' in Notification.prototype

export default class TaskModel extends AbstractModel {
	constructor(data) {
		super(data)

		/** @type {number} */
		this.id = Number(this.id)

		/** @type {string} */
		this.title = this.title?.trim()

		/** @type {string} */
		this.description

		/** @type {boolean} */
		this.done

		/** @type */
		this.doneAt = parseDateOrNull(this.doneAt)

		/** @type {number} */
		this.priority

		/** @type {LabelModel[]} */
		this.labels = this.labels
			.map(l => new LabelModel(l))
			.sort((f, s) => f.title > s.title ? 1 : -1)

		/** @type {UserModel[]} */
		// Parse the assignees into user models
		this.assignees = this.assignees.map(a => {
			return new UserModel(a)
		})

		/** @type {Date} */
		this.dueDate = parseDateOrNull(this.dueDate)

		/** @type {Date} */
		this.startDate = parseDateOrNull(this.startDate)

		/** @type {Date} */
		this.endDate = parseDateOrNull(this.endDate)

		/** @type */
		this.repeatAfter

		// Parse the repeat after into something usable
		this.parseRepeatAfter()

		/** @type {boolean} */
		this.repeatFromCurrentDate

		/** @type {TaskRepeatMode: 0 | 1 | 2} */
		this.repeatMode

		/** @type {Date[]} */
		this.reminderDates = this.reminderDates.map(d => new Date(d))

		// Cancel all scheduled notifications for this task to be sure to only have available notifications
		this.cancelScheduledNotifications().then(() => {
			// Every time we see a reminder, we schedule a notification for it
			this.reminderDates.forEach(d => this.scheduleNotification(d))
		})

		/** @type {number} */
		this.parentTaskId

		/** @type {string} */
		this.hexColor

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		/** @type {number} */
		this.percentDone

		/** @type {{ [relationKind: string]: TaskModel }} */
		this.relatedTasks

		// Make all subtasks to task models
		Object.keys(this.relatedTasks).forEach(relationKind => {
			this.relatedTasks[relationKind] = this.relatedTasks[relationKind].map(t => {
				return new TaskModel(t)
			})
		})

		// Make all attachments to attachment models
		/** @type {AttachmentModel[]} */
		this.attachments = this.attachments.map(a => new AttachmentModel(a))

		/** @type {string} */
		this.identifier

		// Set the task identifier to empty if the list does not have one
		if (this.identifier === `-${this.index}`) {
			this.identifier = ''
		}

		/** @type {number} */
		this.index

		/** @type {boolean} */
		this.isFavorite

		/** @type {SubscriptionModel} */
		this.subscription

		if (typeof this.subscription !== 'undefined' && this.subscription !== null) {
			this.subscription = new SubscriptionModel(this.subscription)
		}

		/** @type {number} */
		this.position

		/** @type {number} */
		this.kanbanPosition

		/** @type {UserModel} */
		this.createdBy = new UserModel(this.createdBy)

		/** @type {Date} */
		this.created = new Date(this.created)

		/** @type {Date} */
		this.updated = new Date(this.updated)

		/** @type {number} */
		this.listId = Number(this.listId)
	}

	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			done: false,
			doneAt: null,
			priority: 0,
			labels: [],
			assignees: [],

			dueDate: 0,
			startDate: 0,
			endDate: 0,
			repeatAfter: 0,
			repeatFromCurrentDate: false,
			repeatMode: REPEAT_MODE_DEFAULT,
			reminderDates: [],
			parentTaskId: 0,
			hexColor: '',
			percentDone: 0,
			relatedTasks: {},
			attachments: [],
			identifier: '',
			index: 0,
			isFavorite: false,
			subscription: null,

			position: 0,
			kanbanPosition: 0,

			createdBy: UserModel,
			created: null,
			updated: null,

			listId: 0, // Meta, only used when creating a new task
		}
	}

	getTextIdentifier() {
		if (this.identifier === '') {
			return `#${this.index}`
		}

		return this.identifier
	}

	getHexColor() {
		if (this.hexColor === '') {
			return `#${this.defaultColor}`
		}

		return this.hexColor
	}

	/////////////////
	// Helper functions
	///////////////

	/**
	 * Parses the "repeat after x seconds" from the task into a usable js object inside the task.
	 * This function should only be called from the constructor.
	 */
	parseRepeatAfter() {
		const repeatAfterHours = (this.repeatAfter / 60) / 60
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
			console.debug('This browser does not support triggered notifications')
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
			console.debug('This browser does not support triggered notifications')
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

