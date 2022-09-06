import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'
import type {ITaskComment} from './ITaskComment'
import type {ITeam} from './ITeam'
import type { IList } from './IList'

export const NOTIFICATION_NAMES = {
	'TASK_COMMENT': 'task.comment',
	'TASK_ASSIGNED': 'task.assigned',
	'TASK_DELETED': 'task.deleted',
	'LIST_CREATED': 'list.created',
	'TEAM_MEMBER_ADDED': 'team.member.added',
} as const

interface Notification {
	doer: IUser
}
interface NotificationTask extends Notification {
	task: ITask
	comment: ITaskComment
}

interface NotificationAssigned extends Notification {
	task: ITask
	assignee: IUser
}

interface NotificationDeleted extends Notification {
	task: ITask
}

interface NotificationCreated extends Notification {
	task: ITask
	list: IList
}

interface NotificationMemberAdded extends Notification {
	member: IUser
	team: ITeam
}

export interface INotification extends IAbstract {
	id: number
	name: string
	notification: NotificationTask | NotificationAssigned | NotificationDeleted | NotificationCreated | NotificationMemberAdded
	read: boolean
	readAt: Date | null

	created: Date
}