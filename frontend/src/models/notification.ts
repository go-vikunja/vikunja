import AbstractModel from './abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, {getDisplayName} from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import ProjectModel from '@/models/project'
import TeamModel from '@/models/team'

import {NOTIFICATION_NAMES, type INotification} from '@/modelTypes/INotification'
import type { IUser } from '@/modelTypes/IUser'
import type { ITeam } from '@/modelTypes/ITeam'
import type { IProject } from '@/modelTypes/IProject'
import type { ITaskComment } from '@/modelTypes/ITaskComment'
import type { ITask } from '@/modelTypes/ITask'

export default class NotificationModel extends AbstractModel<INotification> implements INotification {
	id = 0
	name = ''
	notification!: INotification['notification']
	read = false
	readAt: Date | null = null

	created: Date = new Date()

	constructor(data: Partial<INotification>) {
		super()
		this.assignData(data)

		// Transform raw notification data into proper model instances
		const rawNotification = this.notification as unknown as Record<string, unknown>
		
		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
					comment: new TaskCommentModel(rawNotification.comment as Partial<ITaskComment>),
				}
				break
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
					assignee: new UserModel(rawNotification.assignee as Partial<IUser>),
				}
				break
			case NOTIFICATION_NAMES.TASK_DELETED:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
				}
				break
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
					project: new ProjectModel(rawNotification.project as Partial<IProject>),
				}
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					member: new UserModel(rawNotification.member as Partial<IUser>),
					team: new TeamModel(rawNotification.team as Partial<ITeam>),
				}
				break
			case NOTIFICATION_NAMES.TASK_REMINDER:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
					project: new ProjectModel(rawNotification.project as Partial<IProject>),
				}
				break
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				this.notification = {
					doer: new UserModel(rawNotification.doer as Partial<IUser>),
					task: new TaskModel((rawNotification.task as Partial<ITask>) ?? {}),
				}
				break
		}

		this.created = new Date(this.created)
		this.readAt = this.readAt ? parseDateOrNull(this.readAt as string | Date) : null
	}

	toText(user: IUser | null = null) {
		let who = ''

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				if ('task' in this.notification) {
					return `commented on ${(this.notification.task as TaskModel).getTextIdentifier()}`
				}
				return 'commented on a task'
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
			{
				if ('assignee' in this.notification && 'task' in this.notification) {
					who = `${getDisplayName(this.notification.assignee)}`

					if (user !== null && user.id === this.notification.assignee.id) {
						who = 'you'
					}

					return `assigned ${who} to ${(this.notification.task as TaskModel).getTextIdentifier()}`
				}
				return 'assigned someone to a task'
			}
			case NOTIFICATION_NAMES.TASK_DELETED:
				if ('task' in this.notification) {
					return `deleted ${(this.notification.task as TaskModel).getTextIdentifier()}`
				}
				return 'deleted a task'
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				if ('project' in this.notification) {
					return `created ${this.notification.project.title}`
				}
				return 'created a project'
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
			{
				if ('member' in this.notification && 'team' in this.notification) {
					who = `${getDisplayName(this.notification.member)}`

					if (user !== null && user.id === this.notification.member.id) {
						who = 'you'
					}

					return `added ${who} to the ${this.notification.team.name} team`
				}
				return 'added someone to a team'
			}
			case NOTIFICATION_NAMES.TASK_REMINDER:
			{
				if ('task' in this.notification && 'project' in this.notification) {
					return `Reminder for ${(this.notification.task as TaskModel).getTextIdentifier()} ${this.notification.task.title} (${this.notification.project.title})`
				}
				return 'Task reminder'
			}
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				if ('doer' in this.notification && 'task' in this.notification) {
					return `${getDisplayName(this.notification.doer)} mentioned you on ${(this.notification.task as TaskModel).getTextIdentifier()}`
				}
				return 'Someone mentioned you on a task'
		}

		return ''
	}
}
