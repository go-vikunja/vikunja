// No query module imports, so the query client and every query module can use these without a cycle.
import type {User} from '@/client/generated'
import {ApiError} from '@/client/problem'

export const trashKeys = {
	all: ['trash'] as const,
	tasks: ['trash', 'tasks'] as const,
	taskList: (page: number) => ['trash', 'tasks', page] as const,
	projects: ['trash', 'projects'] as const,
	projectList: (page: number) => ['trash', 'projects', page] as const,
}

export const ERROR_CODE_TASK_IS_IN_TRASH = 4040
export const ERROR_CODE_PROJECT_IS_IN_TRASH = 3018

export type TaskInTrash = {
	project_id: number
	deleted_at: string
	deleted_by?: User
}

export type ProjectInTrash = {
	id: number
	deleted_at: string
	deleted_by?: User
}

function trashDetailsFromError(cause: unknown, errorCode: number): unknown {
	if (!(cause instanceof ApiError) || cause.code !== errorCode) return null
	return cause.details ?? null
}

export function taskInTrashFromError(cause: unknown): TaskInTrash | null {
	return trashDetailsFromError(cause, ERROR_CODE_TASK_IS_IN_TRASH) as TaskInTrash | null
}

export function projectInTrashFromError(cause: unknown): ProjectInTrash | null {
	return trashDetailsFromError(cause, ERROR_CODE_PROJECT_IS_IN_TRASH) as ProjectInTrash | null
}

// Views render these in place instead of a toast
export function isInTrashError(cause: unknown): boolean {
	return taskInTrashFromError(cause) !== null || projectInTrashFromError(cause) !== null
}
