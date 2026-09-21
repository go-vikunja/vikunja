import type {DatabaseNotification, User, Task, Project, Team} from '@/client/generated'
import type {RouteLocationRaw} from 'vue-router'
import {getDisplayName} from '@/helpers/user'
import {getTaskIdentifier} from '@/helpers/task'

export const NOTIFICATION_NAMES = {
	'TASK_COMMENT': 'task.comment',
	'TASK_ASSIGNED': 'task.assigned',
	'TASK_DELETED': 'task.deleted',
	'TASK_CREATED': 'task.created',
	'TASK_REMINDER': 'task.reminder',
	'PROJECT_CREATED': 'project.created',
	'TEAM_MEMBER_ADDED': 'team.member.added',
	'TASK_MENTIONED': 'task.mentioned',
} as const

function record(value: unknown): Record<string, unknown> {
	return typeof value === 'object' && value !== null ? value as Record<string, unknown> : {}
}

export function notificationDoer(n: DatabaseNotification): User | undefined {
	const doer = record(record(n.notification).doer)
	return typeof doer.id === 'number' ? doer as User : undefined
}

export function notificationText(n: DatabaseNotification, user: Pick<User, 'id'> | null = null): string {
	const payload = record(n.notification)
	const task = record(payload.task) as Task
	const project = record(payload.project) as Project
	const assignee = record(payload.assignee) as User
	const member = record(payload.member) as User
	const team = record(payload.team) as Team
	const doer = record(payload.doer) as User
	let who: string
	switch (n.name) {
		case NOTIFICATION_NAMES.TASK_COMMENT:
			return `commented on ${getTaskIdentifier(task)}`
		case NOTIFICATION_NAMES.TASK_ASSIGNED:
			who = `${getDisplayName(assignee)}`

			if (user !== null && user.id === assignee.id) {
				who = 'you'
			}

			return `assigned ${who} to ${getTaskIdentifier(task)}`
		case NOTIFICATION_NAMES.TASK_DELETED:
			return `deleted ${getTaskIdentifier(task)}`
		case NOTIFICATION_NAMES.TASK_CREATED:
			return `created ${getTaskIdentifier(task)}`
		case NOTIFICATION_NAMES.PROJECT_CREATED:
			return `created ${project.title}`
		case NOTIFICATION_NAMES.TEAM_MEMBER_ADDED:
			who = `${getDisplayName(member)}`

			if (user !== null && user.id === member.id) {
				who = 'you'
			}

			return `added ${who} to the ${team.name} team`
		case NOTIFICATION_NAMES.TASK_REMINDER:
			return `Reminder for ${getTaskIdentifier(task)} ${task.title} (${project.title})`
		case NOTIFICATION_NAMES.TASK_MENTIONED:
			return `${getDisplayName(doer)} mentioned you on ${getTaskIdentifier(task)}`
	}

	return ''
}

export function notificationRoute(n: DatabaseNotification): RouteLocationRaw | null {
	const payload = record(n.notification)
	if (['task.comment', 'task.assigned', 'task.reminder', 'task.mentioned', 'task.created'].includes(n.name ?? '')) {
		const id = record(payload.task).id
		return typeof id === 'number' && id > 0 ? {name: 'task.detail', params: {id}} : null
	}
	if (n.name === NOTIFICATION_NAMES.PROJECT_CREATED) {
		const projectId = record(payload.project).id
		return typeof projectId === 'number' && projectId > 0 ? {name: 'task.index', params: {projectId}} : null
	}
	if (n.name === NOTIFICATION_NAMES.TEAM_MEMBER_ADDED) {
		const id = record(payload.team).id
		return typeof id === 'number' && id > 0 ? {name: 'teams.edit', params: {id}} : null
	}
	return null
}
