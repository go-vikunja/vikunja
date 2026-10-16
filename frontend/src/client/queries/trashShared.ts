// No query module imports, so the query client and every query module can use these without a cycle.
import type {User} from '@/client/generated'

export const trashKeys = {
	all: ['trash'] as const,
	tasks: ['trash', 'tasks'] as const,
	taskList: (page: number) => ['trash', 'tasks', page] as const,
	projects: ['trash', 'projects'] as const,
	projectList: (page: number) => ['trash', 'projects', page] as const,
}

export const ERROR_CODE_TASK_IS_IN_TRASH = 4036
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
	if (typeof cause !== 'object' || cause === null) return null
	const {code, details} = cause as {code?: number, details?: unknown}
	if (code !== errorCode || typeof details !== 'object' || details === null) return null
	return details
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
