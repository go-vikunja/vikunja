import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, {getDisplayName} from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import ProjectModel from '@/models/project'
import TeamModel from '@/models/team'

import {NOTIFICATION_NAMES, type INotification} from '@/modelTypes/INotification'
import type { IUser } from '@/modelTypes/IUser'
import type { ITask } from '@/modelTypes/ITask'

export default class NotificationModel extends AbstractModel<INotification> implements INotification {
	id = 0
	name = ''
	notification: INotification['notification'] = {} as INotification['notification']
	read = false
	readAt: Date | null = null

	created: Date = new Date()

	constructor(data: Partial<INotification> = {}) {
		super()
		this.assignData(data)

		// Only process notification if it exists
		if (this.notification && typeof this.notification === 'object') {
			switch (this.name) {
				case NOTIFICATION_NAMES.TASK_COMMENT:
					if ('doer' in this.notification && 'task' in this.notification && 'comment' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							task: new TaskModel(this.notification.task),
							comment: new TaskCommentModel(this.notification.comment),
						}
					}
					break
				case NOTIFICATION_NAMES.TASK_ASSIGNED:
					if ('doer' in this.notification && 'task' in this.notification && 'assignee' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							task: new TaskModel(this.notification.task),
							assignee: new UserModel(this.notification.assignee),
						}
					}
					break
				case NOTIFICATION_NAMES.TASK_DELETED:
					if ('doer' in this.notification && 'task' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							task: new TaskModel(this.notification.task),
						}
					}
					break
				case NOTIFICATION_NAMES.PROJECT_CREATED:
					if ('doer' in this.notification && 'project' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							task: new TaskModel(), // Required by interface
							project: new ProjectModel(this.notification.project),
						}
					}
					break
				case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
					if ('doer' in this.notification && 'member' in this.notification && 'team' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							member: new UserModel(this.notification.member),
							team: new TeamModel({...this.notification.team, createdBy: this.notification.team.createdBy || new UserModel()}),
						}
					}
					break
				case NOTIFICATION_NAMES.TASK_REMINDER:
					if ('task' in this.notification && 'project' in this.notification) {
						this.notification = {
							doer: new UserModel(),
							task: new TaskModel(this.notification.task),
							project: new ProjectModel(this.notification.project),
						}
					}
					break
				case NOTIFICATION_NAMES.TASK_MENTIONED:
					if ('doer' in this.notification && 'task' in this.notification) {
						this.notification = {
							doer: new UserModel(this.notification.doer),
							task: new TaskModel(this.notification.task),
						}
					}
					break
			}
		}

		this.created = this.created ? new Date(this.created) : new Date()
		this.readAt = parseDateOrNull(this.readAt as string | Date)
	}

	toText(user: IUser | null = null) {
		let who = ''

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				if ('task' in this.notification && this.notification.task) {
					const task = this.notification.task as ITask & {getTextIdentifier(): string}
					return `commented on ${task.getTextIdentifier()}`
				}
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				if ('assignee' in this.notification && 'task' in this.notification && this.notification.task) {
					who = `${getDisplayName(this.notification.assignee)}`

					if (user !== null && user.id === this.notification.assignee.id) {
						who = 'you'
					}

					const task = this.notification.task as ITask & {getTextIdentifier(): string}
					return `assigned ${who} to ${task.getTextIdentifier()}`
				}
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				if ('task' in this.notification && this.notification.task) {
					const task = this.notification.task as ITask & {getTextIdentifier(): string}
					return `deleted ${task.getTextIdentifier()}`
				}
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				if ('project' in this.notification) {
					return `created ${this.notification.project.title}`
				}
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				if ('member' in this.notification && 'team' in this.notification) {
					who = `${getDisplayName(this.notification.member)}`

					if (user !== null && user.id === this.notification.member.id) {
						who = 'you'
					}

					return `added ${who} to the ${this.notification.team.name} team`
				}
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				if ('task' in this.notification && 'project' in this.notification && this.notification.task) {
					const task = this.notification.task as ITask & {getTextIdentifier(): string}
					return `Reminder for ${task.getTextIdentifier()} ${task.title} (${this.notification.project.title})`
				}
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				if ('doer' in this.notification && 'task' in this.notification && this.notification.task) {
					const task = this.notification.task as ITask & {getTextIdentifier(): string}
					return `${getDisplayName(this.notification.doer)} mentioned you on ${task.getTextIdentifier()}`
				}
				break
		}

		return ''
	}
}
