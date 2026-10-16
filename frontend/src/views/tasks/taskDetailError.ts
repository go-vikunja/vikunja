import {isRequestContextAbort} from '@/client/requestContext'
import {taskInTrashFromError} from '@/client/queries/trash'

export type TaskLoadErrorAction = 'ignore' | 'inTrash' | 'notFound' | 'report'

export function taskLoadErrorAction(cause: unknown): TaskLoadErrorAction {
	if (isRequestContextAbort(cause)) return 'ignore'
	if (taskInTrashFromError(cause)) return 'inTrash'
	if ([403, 404].includes((cause as {status?: number} | null)?.status ?? 0)) return 'notFound'
	return 'report'
}
