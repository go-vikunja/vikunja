// No query module imports: the query client and projects.ts need these, and trash.ts imports projects.ts.
import type {User} from '@/client/generated'
import {ApiError} from '@/client/problem'

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
	can_restore: boolean
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
