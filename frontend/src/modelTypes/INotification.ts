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

interface NotificationTaskComment extends Notification {
	task: ITask
	comment: ITaskComment
}

interface NotificationTask extends Notification {
	task: ITask
}

interface NotificationAssigned extends Notification {
	task: ITask
	assignee: IUser
}

interface NotificationCreated extends Notification {
	task: ITask
	project: IProject
}

interface NotificationTaskReminder extends Notification {
	task: ITask
	project: IProject
}

interface NotificationMemberAdded extends Notification {
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
