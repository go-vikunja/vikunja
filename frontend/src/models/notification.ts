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
	notification: INotification['notification'] = {} as any
	read = false
	readAt: Date | null = null

	created: Date = new Date()

	constructor(data: Partial<INotification>) {
		super()
		this.assignData(data)

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
						comment: new TaskCommentModel(notification.comment),
					}
				}
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
						assignee: new UserModel(notification.assignee),
					}
				}
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
					}
				}
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
						project: new ProjectModel(notification.project),
					} as any
				}
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						member: new UserModel(notification.member),
						team: new TeamModel(notification.team) as any,
					}
				}
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
						project: new ProjectModel(notification.project),
					} as any
				}
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				{
					const notification = this.notification as any
					this.notification = {
						doer: new UserModel(notification.doer),
						task: new TaskModel(notification.task),
					}
				}
				break
		}

		this.created = new Date(this.created)
		this.readAt = parseDateOrNull(this.readAt as any)
	}

	toText(user: IUser | null = null) {
		let who = ''

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				return `commented on ${(this.notification as any).task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				{
					const notification = this.notification as any
					who = `${getDisplayName(notification.assignee)}`

					if (user !== null && user.id === notification.assignee.id) {
						who = 'you'
					}

					return `assigned ${who} to ${notification.task.getTextIdentifier()}`
				}
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${(this.notification as any).task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				return `created ${(this.notification as any).project.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				{
					const notification = this.notification as any
					who = `${getDisplayName(notification.member)}`

					if (user !== null && user.id === notification.member.id) {
						who = 'you'
					}

					return `added ${who} to the ${notification.team.name} team`
				}
			case NOTIFICATION_NAMES.TASK_REMINDER:
				{
					const notification = this.notification as any
					return `Reminder for ${notification.task.getTextIdentifier()} ${notification.task.title} (${notification.project.title})`
				}
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				return `${getDisplayName((this.notification as any).doer)} mentioned you on ${(this.notification as any).task.getTextIdentifier()}`
		}

		return ''
	}
}
