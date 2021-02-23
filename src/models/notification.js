import AbstractModel from '@/models/abstractModel'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'
import UserModel from '@/models/user'
import TaskModel from '@/models/task'
import TaskCommentModel from '@/models/taskComment'
import ListModel from '@/models/list'
import TeamModel from '@/models/team'
import names from './notificationNames.json'

export default class NotificationModel extends AbstractModel {
	constructor(data) {
		super(data)

		switch (this.name) {
			case names.TASK_COMMENT:
				this.notification.doer = new UserModel(this.notification.doer)
				this.notification.task = new TaskModel(this.notification.task)
				this.notification.comment = new TaskCommentModel(this.notification.comment)
				break
			case names.TASK_ASSIGNED:
				this.notification.doer = new UserModel(this.notification.doer)
				this.notification.task = new TaskModel(this.notification.task)
				this.notification.assignee = new UserModel(this.notification.assignee)
				break
			case names.TASK_DELETED:
				this.notification.doer = new UserModel(this.notification.doer)
				this.notification.task = new TaskModel(this.notification.task)
				break
			case names.LIST_CREATED:
				this.notification.doer = new UserModel(this.notification.doer)
				this.notification.list = new ListModel(this.notification.list)
				break
			case names.TEAM_MEMBER_ADDED:
				this.notification.doer = new UserModel(this.notification.doer)
				this.notification.member = new UserModel(this.notification.member)
				this.notification.team = new TeamModel(this.notification.team)
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
			case names.TASK_COMMENT:
				return `commented on ${this.notification.task.getTextIdentifier()}`
			case names.TASK_ASSIGNED:
				who = `${this.notification.assignee.getDisplayName()}`

				if (user !== null && user.id === this.notification.assignee.id) {
					who = 'you'
				}

				return `assigned ${who} to ${this.notification.task.getTextIdentifier()}`
			case names.TASK_DELETED:
				return `deleted ${this.notification.task.getTextIdentifier()}`
			case names.LIST_CREATED:
				return `created ${this.notification.list.title}`
			case names.TEAM_MEMBER_ADDED:
				who = `${this.notification.member.getDisplayName()}`

				if (user !== null && user.id === this.notification.member.id) {
					who = 'you'
				}

				return `added ${who} to the ${this.notification.team.name} team`
		}

		return ''
	}
}
