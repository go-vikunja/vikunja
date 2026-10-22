import {isRequestContextAbort} from '@/client/requestContext'

export type TaskLoadErrorAction = 'ignore' | 'notFound' | 'report'

export function taskLoadErrorAction(cause: unknown): TaskLoadErrorAction {
	if (isRequestContextAbort(cause)) return 'ignore'
	if ([403, 404].includes((cause as {status?: number} | null)?.status ?? 0)) return 'notFound'
	return 'report'
}
