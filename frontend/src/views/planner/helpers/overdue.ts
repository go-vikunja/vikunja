import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'

export function overdueCutoff(): dayjs.Dayjs {
	return dayjs().startOf('day')
}

/**
 * A task is overdue when it is not done and its planner anchor lies before
 * today: a scheduled block that ended in the past, a start-only task whose
 * start day passed, or a due-only task whose due date passed. A task with a
 * schedule reaching into today or the future is never overdue here even if
 * its due date already passed — it is planned, so rescheduling resolved it.
 */
export function isOverdue(task: ITask, cutoff: dayjs.Dayjs = overdueCutoff()): boolean {
	if (task.done) {
		return false
	}
	if (task.endDate) {
		return dayjs(task.endDate).isBefore(cutoff)
	}
	if (task.startDate) {
		return dayjs(task.startDate).isBefore(cutoff)
	}
	if (task.dueDate) {
		return dayjs(task.dueDate).isBefore(cutoff)
	}
	return false
}

// Sort key for the overdue list: the date that made the task overdue.
export function overdueAnchor(task: ITask): Date | null {
	return task.dueDate ?? task.endDate ?? task.startDate ?? null
}
