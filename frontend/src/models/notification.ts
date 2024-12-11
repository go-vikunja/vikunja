import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, {getDisplayName} from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import ProjectModel from '@/models/project'
import TeamModel from '@/models/team'

import {NOTIFICATION_NAMES, type INotification} from '@/modelTypes/INotification'
import type { IUser } from '@/modelTypes/IUser'

export default class NotificationModel extends AbstractModel<INotification> implements INotification {
	id = 0
	name = ''
	notification: INotification['notification'] = null
	read = false
	readAt: Date | null = null

	created: Date

	constructor(data: Partial<INotification>) {
		super()
		this.assignData(data)

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					task: new TaskModel(this.notification.task),
					comment: new TaskCommentModel(this.notification.comment),
				}
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					task: new TaskModel(this.notification.task),
					assignee: new UserModel(this.notification.assignee),
				}
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					task: new TaskModel(this.notification.task),
				}
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					project: new ProjectModel(this.notification.project),
				}
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					member: new UserModel(this.notification.member),
					team: new TeamModel(this.notification.team),
				}
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				this.notification = {
					task: new TaskModel(this.notification.task),
					project: new ProjectModel(this.notification.project),
				}
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					task: new TaskModel(this.notification.task),
				}
				break
		}

		this.created = new Date(this.created)
		this.readAt = parseDateOrNull(this.readAt)
	}

	toText(user: IUser | null = null) {
		let who = ''

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				return `commented on ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				who = `${getDisplayName(this.notification.assignee)}`

				if (user !== null && user.id === this.notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				return `created ${this.notification.project.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				who = `${getDisplayName(this.notification.member)}`

				if (user !== null && user.id === this.notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${this.notification.team.name} team`
			case NOTIFICATION_NAMES.TASK_REMINDER:
				return `Reminder for ${this.notification.task.getTextIdentifier()} ${this.notification.task.title} (${this.notification.project.title})`
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				return `${getDisplayName(this.notification.doer)} mentioned you on ${this.notification.task.getTextIdentifier()}`
		}

		return ''
	}
}
