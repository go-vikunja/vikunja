import {isRequestContextAbort} from '@/client/requestContext'
import {
	projectInTrashFromError,
	taskInTrashFromError,
	type ProjectInTrash,
	type TaskInTrash,
} from '@/client/queries/trashErrors'

export type TaskLoadErrorAction =
	| {action: 'inTrash', inTrash: TaskInTrash}
	| {action: 'projectInTrash', inTrash: ProjectInTrash}
	| {action: 'ignore' | 'notFound' | 'report'}

export function taskLoadErrorAction(cause: unknown): TaskLoadErrorAction {
	if (isRequestContextAbort(cause)) return {action: 'ignore'}
	const inTrash = taskInTrashFromError(cause)
	if (inTrash) return {action: 'inTrash', inTrash}
	const projectInTrash = projectInTrashFromError(cause)
	if (projectInTrash) return {action: 'projectInTrash', inTrash: projectInTrash}
	if ([403, 404].includes((cause as {status?: number} | null)?.status ?? 0)) return {action: 'notFound'}
	return {action: 'report'}
}
