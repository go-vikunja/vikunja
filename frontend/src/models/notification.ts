import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import {getDisplayName} from '@/helpers/user'
import {getTaskIdentifier} from '@/helpers/task'
import type {Task} from '@/client/generated'
import type {TaskComment} from '@/client/generated'
import type {Team} from '@/client/generated'
import {objectToSnakeCase} from '@/helpers/case'

import {NOTIFICATION_NAMES, type INotification} from '@/modelTypes/INotification'
import type {User as IUser} from '@/client/generated'

type NotificationData = {
	doer: IUser
	task: Task
	comment: TaskComment
	assignee: IUser
	project: Extract<INotification['notification'], {project: unknown}>['project']
	member: IUser
	team: Team
}

function asNotificationPayload(data: Partial<NotificationData>): INotification['notification'] {
	return data as unknown as INotification['notification']
}

export default class NotificationModel extends AbstractModel<INotification> implements INotification {
	id = 0
	name = ''
	notification = null as unknown as INotification['notification']
	read = false
	readAt: Date | null = null

	created!: Date

	constructor(data: Partial<INotification>) {
		super()
		const task = data.notification && 'task' in data.notification ? data.notification.task : undefined
		const comment = data.notification && 'comment' in data.notification ? data.notification.comment : undefined
		this.assignData(data)
		const notification = data.notification as unknown as NotificationData

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					task,
					comment,
				})
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					task,
					assignee: notification.assignee,
				})
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					task,
				})
				break
			case NOTIFICATION_NAMES.TASK_CREATED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					task,
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					member: notification.member,
					team: objectToSnakeCase(notification.team) as Team,
				})
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				this.notification = asNotificationPayload({
					task,
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				this.notification = asNotificationPayload({
					doer: notification.doer,
					task,
				})
				break
		}

		this.created = new Date(this.created)
		this.readAt = this.readAt === null ? null : parseDateOrNull(this.readAt)
	}

	toText(user: Pick<IUser, 'id'> | null = null) {
		let who: string
		const notification = this.notification as unknown as NotificationData

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				return `commented on ${getTaskIdentifier(notification.task)}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				who = `${getDisplayName(notification.assignee)}`

				if (user !== null && user.id === notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${getTaskIdentifier(notification.task)}`
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${getTaskIdentifier(notification.task)}`
			case NOTIFICATION_NAMES.TASK_CREATED:
				return `created ${getTaskIdentifier(notification.task)}`
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				return `created ${notification.project.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				who = `${getDisplayName(notification.member)}`

				if (user !== null && user.id === notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${notification.team.name} team`
			case NOTIFICATION_NAMES.TASK_REMINDER:
				return `Reminder for ${getTaskIdentifier(notification.task)} ${notification.task.title} (${notification.project.title})`
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				return `${getDisplayName(notification.doer)} mentioned you on ${getTaskIdentifier(notification.task)}`
		}

		return ''
	}
}
