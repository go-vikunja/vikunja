import AbstractModel, { type IAbstract } from '@/models/abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel, { type IUser } from '@/models/user'
import TaskModel, { type ITask } from '@/models/task'
import TaskCommentModel, { type ITaskComment } from '@/models/taskComment'
import ListModel from '@/models/list'
import TeamModel, { type ITeam } from '@/models/team'

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

export default class NotificationModel extends AbstractModel implements INotification {
	id!: number
	name!: string
	notification!: NotificationTask | NotificationAssigned | NotificationDeleted | NotificationCreated | NotificationMemberAdded
	read!: boolean
	readAt: Date | null

	created: Date

	constructor(data) {
		super(data)

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
			case NOTIFICATION_NAMES.LIST_CREATED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					list: new ListModel(this.notification.list),
				}
				break
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				this.notification = {
					doer: new UserModel(this.notification.doer),
					member: new UserModel(this.notification.member),
					team: new TeamModel(this.notification.team),
				}
				break
		}

		this.created = new Date(this.created)
		this.readAt = parseDateOrNull(this.readAt)
	}

	defaults() {
		return {
			id: 0,
			name: '',
			notification: null,
			read: false,
			readAt: null,
		}
	}

	toText(user = null) {
		let who = ''

		switch (this.name) {
			case NOTIFICATION_NAMES.TASK_COMMENT:
				return `commented on ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_ASSIGNED:
				who = `${this.notification.assignee.getDisplayName()}`

				if (user !== null && user.id === this.notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.TASK_DELETED:
				return `deleted ${this.notification.task.getTextIdentifier()}`
			case NOTIFICATION_NAMES.LIST_CREATED:
				return `created ${this.notification.list.title}`
			case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
				who = `${this.notification.member.getDisplayName()}`

				if (user !== null && user.id === this.notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${this.notification.team.name} team`
		}

		return ''
	}
}
