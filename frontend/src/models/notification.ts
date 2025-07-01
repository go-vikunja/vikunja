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
				return `commented on ${(this.notification as {task: TaskModel}).task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
			{
				const notification = this.notification as {assignee: UserModel, task: TaskModel}
				who = `${getDisplayName(notification.assignee)}`

				if (user !== null && user.id === notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${notification.task.getTextIdentifier()}`
			}
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${(this.notification as {task: TaskModel}).task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.PROJECT_CREATED:
				return `created ${(this.notification as {project: ProjectModel}).project.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
			{
				const notification = this.notification as {member: UserModel, team: TeamModel}
				who = `${getDisplayName(notification.member)}`

				if (user !== null && user.id === notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${notification.team.name} team`
			}
			case NOTIFICATION_NAMES.TASK_REMINDER:
			{
				const notification = this.notification as {task: TaskModel, project: ProjectModel}
				return `Reminder for ${notification.task.getTextIdentifier()} ${notification.task.title} (${notification.project.title})`
			}
			case NOTIFICATION_NAMES.TASK_MENTIONED:
				return `${getDisplayName((this.notification as {doer: UserModel}).doer)} mentioned you on ${(this.notification as {task: TaskModel}).task.getTextIdentifier()}`
		}

		return ''
	}
}
