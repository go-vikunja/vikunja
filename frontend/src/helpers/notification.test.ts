import {describe, expect, it} from 'vitest'
import type {DatabaseNotification} from '@/client/generated'
import {
	NOTIFICATION_NAMES,
	notificationDoer,
	notificationRoute,
	notificationText,
} from './notification'

const TASK = {
	id: 3,
	identifier: '#12',
	index: 12,
	title: 'Write docs',
}
const PROJECT = {
	id: 4,
	title: 'Docs',
}
const TEAM = {
	id: 5,
	name: 'Design',
}
const SAM = {
	id: 7,
	username: 'sam',
}
const ANA = {
	id: 9,
	username: 'ana',
}

function notification(name: string, payload: unknown): DatabaseNotification {
	return {
		id: 1,
		name,
		notification: payload,
	}
}

describe('notificationText', () => {
	it('names the task for comment, delete and create notifications', () => {
		expect(notificationText(notification(NOTIFICATION_NAMES.TASK_COMMENT, {task: TASK}))).toBe('commented on #12')
		expect(notificationText(notification(NOTIFICATION_NAMES.TASK_DELETED, {task: TASK}))).toBe('deleted #12')
		expect(notificationText(notification(NOTIFICATION_NAMES.TASK_CREATED, {task: TASK}))).toBe('created #12')
	})

	it('addresses the assignee as you only when they are the current user', () => {
		const assigned = notification(NOTIFICATION_NAMES.TASK_ASSIGNED, {
			task: TASK,
			assignee: SAM,
		})
		expect(notificationText(assigned)).toBe('assigned sam to #12')
		expect(notificationText(assigned, {id: 9})).toBe('assigned sam to #12')
		expect(notificationText(assigned, {id: 7})).toBe('assigned you to #12')
	})

	it('addresses the new team member as you only when they are the current user', () => {
		const added = notification(NOTIFICATION_NAMES.TEAM_MEMBER_ADDED, {
			member: SAM,
			team: TEAM,
		})
		expect(notificationText(added)).toBe('added sam to the Design team')
		expect(notificationText(added, {id: 9})).toBe('added sam to the Design team')
		expect(notificationText(added, {id: 7})).toBe('added you to the Design team')
	})

	it('names the project for a created project', () => {
		const created = notification(NOTIFICATION_NAMES.PROJECT_CREATED, {project: PROJECT})
		expect(notificationText(created)).toBe('created Docs')
	})

	it('spells out task and project for a reminder', () => {
		const reminder = notification(NOTIFICATION_NAMES.TASK_REMINDER, {
			task: TASK,
			project: PROJECT,
		})
		expect(notificationText(reminder)).toBe('Reminder for #12 Write docs (Docs)')
	})

	it('names the doer for a mention', () => {
		const mentioned = notification(NOTIFICATION_NAMES.TASK_MENTIONED, {
			task: TASK,
			doer: ANA,
		})
		expect(notificationText(mentioned)).toBe('ana mentioned you on #12')
		expect(notificationText(mentioned, {id: 9})).toBe('ana mentioned you on #12')
	})

	it('returns an empty string for an unknown name or a missing payload', () => {
		expect(notificationText(notification('task.exploded', {task: TASK}))).toBe('')
		expect(notificationText(notification(NOTIFICATION_NAMES.TASK_COMMENT, null))).toBe('commented on #0')
	})
})

describe('notificationRoute', () => {
	it('routes every task notification to the task detail', () => {
		const taskNames = [
			NOTIFICATION_NAMES.TASK_COMMENT,
			NOTIFICATION_NAMES.TASK_ASSIGNED,
			NOTIFICATION_NAMES.TASK_REMINDER,
			NOTIFICATION_NAMES.TASK_MENTIONED,
			NOTIFICATION_NAMES.TASK_CREATED,
		]
		for (const name of taskNames) {
			expect(notificationRoute(notification(name, {task: TASK}))).toEqual({
				name: 'task.detail',
				params: {id: 3},
			})
		}
	})

	it('routes a created project and a team membership', () => {
		expect(notificationRoute(notification(NOTIFICATION_NAMES.PROJECT_CREATED, {project: PROJECT}))).toEqual({
			name: 'task.index',
			params: {projectId: 4},
		})
		expect(notificationRoute(notification(NOTIFICATION_NAMES.TEAM_MEMBER_ADDED, {team: TEAM}))).toEqual({
			name: 'teams.edit',
			params: {id: 5},
		})
	})

	it('has no route without a usable target', () => {
		expect(notificationRoute(notification('task.exploded', {task: TASK}))).toBeNull()
		expect(notificationRoute(notification(NOTIFICATION_NAMES.TASK_COMMENT, {task: {id: 0}}))).toBeNull()
		expect(notificationRoute(notification(NOTIFICATION_NAMES.TASK_COMMENT, {}))).toBeNull()
		expect(notificationRoute(notification(NOTIFICATION_NAMES.PROJECT_CREATED, {}))).toBeNull()
		expect(notificationRoute(notification(NOTIFICATION_NAMES.TEAM_MEMBER_ADDED, {}))).toBeNull()
		expect(notificationRoute({})).toBeNull()
	})
})

describe('notificationDoer', () => {
	it('returns the doer only when the payload carries one with an id', () => {
		expect(notificationDoer(notification(NOTIFICATION_NAMES.TASK_COMMENT, {doer: ANA}))).toEqual(ANA)
		const withoutId = notification(NOTIFICATION_NAMES.TASK_COMMENT, {doer: {username: 'ana'}})
		expect(notificationDoer(withoutId)).toBeUndefined()
		expect(notificationDoer(notification(NOTIFICATION_NAMES.TASK_COMMENT, {}))).toBeUndefined()
		expect(notificationDoer({})).toBeUndefined()
	})
})
