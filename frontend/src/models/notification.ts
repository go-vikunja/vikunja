import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, {getDisplayName} from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import TeamModel from '@/models/team'

import {NOTIFICATION_NAMES, type INotification} from '@/modelTypes/INotification'
import type {IUser} from '@/modelTypes/IUser'

type NotificationData = {
	doer: UserModel
	task: TaskModel
	comment: TaskCommentModel
	assignee: UserModel
	project: Extract<INotification['notification'], {project: unknown}>['project']
	member: UserModel
	team: TeamModel
}

type RawNotificationData = Omit<NotificationData, 'team'> & {
	team: ConstructorParameters<typeof TeamModel>[0]
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
		this.assignData(data)
		const notification = this.notification as unknown as RawNotificationData

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					task: new TaskModel(notification.task),
					comment: new TaskCommentModel(notification.comment),
				})
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					task: new TaskModel(notification.task),
					assignee: new UserModel(notification.assignee),
				})
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					task: new TaskModel(notification.task),
				})
				break
			case NOTIFICATION_NAMES.TASK_CREATED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					task: new TaskModel(notification.task),
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					member: new UserModel(notification.member),
					team: new TeamModel(notification.team),
				})
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				this.notification = asNotificationPayload({
					task: new TaskModel(notification.task),
					project: notification.project,
				})
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				this.notification = asNotificationPayload({
					doer: new UserModel(notification.doer),
					task: new TaskModel(notification.task),
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
				return `commented on ${notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				who = `${getDisplayName(notification.assignee)}`

				if (user !== null && user.id === notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_CREATED:
				return `created ${notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				return `created ${notification.project.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				who = `${getDisplayName(notification.member)}`

				if (user !== null && user.id === notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${notification.team.name} team`
			case NOTIFICATION_NAMES.TASK_REMINDER:
				return `Reminder for ${notification.task.getTextIdentifier()} ${notification.task.title} (${notification.project.title})`
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				return `${getDisplayName(notification.doer)} mentioned you on ${notification.task.getTextIdentifier()}`
		}

		return ''
	}
}
