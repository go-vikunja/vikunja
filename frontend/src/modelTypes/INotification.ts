import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'
import type {ITaskComment} from './ITaskComment'
import type {ITeam} from './ITeam'
import type { IProject } from './IProject'

export const NOTIFICATION_NAMES = {
	'TASK_COMMENT': 'task.comment',
	'TASK_ASSIGNED': 'task.assigned',
	'TASK_DELETED': 'task.deleted',
	'TASK_REMINDER': 'task.reminder',
	'PROJECT_CREATED': 'project.created',
	'TEAM_MEMBER_ADDED': 'team.member.added',
	'TASK_MENTIONED': 'task.mentioned',
} as const

interface Notification {
	doer: IUser
}

export interface NotificationTaskComment extends Notification {
	task: ITask
	comment: ITaskComment
}

export interface NotificationTask extends Notification {
	task: ITask
}

export interface NotificationAssigned extends Notification {
	task: ITask
	assignee: IUser
}

export interface NotificationCreated extends Notification {
	task: ITask
	project: IProject
}

export interface NotificationTaskReminder extends Notification {
	task: ITask
	project: IProject
}

export interface NotificationMemberAdded extends Notification {
	member: IUser
	team: ITeam
}

export interface INotification extends IAbstract {
	id: number
	name: string
	notification: NotificationTaskComment | NotificationTask | NotificationAssigned | NotificationCreated | NotificationMemberAdded | NotificationTaskReminder
	read: boolean
	readAt: Date | null

	created: Date
}
