import {isRequestContextAbort} from '@/client/requestContext'
import {taskInTrashFromError, type TaskInTrash} from '@/client/queries/trash'

export type TaskLoadErrorAction =
	| {action: 'inTrash', inTrash: TaskInTrash}
	| {action: 'ignore' | 'notFound' | 'report'}

export function taskLoadErrorAction(cause: unknown): TaskLoadErrorAction {
	if (isRequestContextAbort(cause)) return {action: 'ignore'}
	const inTrash = taskInTrashFromError(cause)
	if (inTrash) return {action: 'inTrash', inTrash}
	if ([403, 404].includes((cause as {status?: number} | null)?.status ?? 0)) return {action: 'notFound'}
	return {action: 'report'}
}
