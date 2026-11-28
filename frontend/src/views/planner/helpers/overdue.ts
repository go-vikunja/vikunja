import dayjs from 'dayjs'

import type {TaskResponse} from '@/client/queries/tasks'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

export function overdueCutoff(): dayjs.Dayjs {
	return dayjs().startOf('day')
}

// Sort key and date label for the overdue list, and the date isOverdue judges:
// a scheduled block's end, else a start-only task's start, else a due date.
export function overdueAnchor(task: TaskResponse): Date | null {
	return parseDateOrNull(task.end_date)
		?? parseDateOrNull(task.start_date)
		?? parseDateOrNull(task.due_date)
}

/**
 * A task is overdue when it is not done and its planner anchor lies before
 * today: a scheduled block that ended in the past, a start-only task whose
 * start day passed, or a due-only task whose due date passed. A task with a
 * schedule reaching into today or the future is never overdue here even if
 * its due date already passed — it is planned, so rescheduling resolved it.
 */
export function isOverdue(task: TaskResponse, cutoff: dayjs.Dayjs = overdueCutoff()): boolean {
	if (task.done) {
		return false
	}
	const anchor = overdueAnchor(task)
	return anchor !== null && dayjs(anchor).isBefore(cutoff)
}
