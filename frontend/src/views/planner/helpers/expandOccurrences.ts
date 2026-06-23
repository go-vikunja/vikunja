import dayjs, {type ManipulateType} from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import type {IRepeatAfter} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {parseRepeatAfter} from '@/models/task'
import type {PlannedOccurrence} from './types'

// Guard against pathological repeat intervals projecting forever.
const MAX_OCCURRENCES = 366

const TYPE_TO_DAYJS: Record<IRepeatAfter['type'], ManipulateType> = {
	seconds: 'second',
	minutes: 'minute',
	hours: 'hour',
	days: 'day',
	weeks: 'week',
	months: 'month',
	years: 'year',
}

function getRepeatStep(task: ITask): {amount: number, unit: ManipulateType} | null {
	// Monthly mode repeats on the same day each month regardless of repeatAfter.
	if (task.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH) {
		return {amount: 1, unit: 'month'}
	}

	const repeat: IRepeatAfter = typeof task.repeatAfter === 'number'
		? parseRepeatAfter(task.repeatAfter)
		: task.repeatAfter

	if (!repeat || repeat.amount <= 0) {
		return null
	}

	return {amount: repeat.amount, unit: TYPE_TO_DAYJS[repeat.type]}
}

/**
 * Projects a timed task's occurrences across [from, to].
 *
 * The stored task itself (at its current start/end) is the only real,
 * editable instance; every projected future occurrence is a read-only ghost.
 * Projection is keyed off the task's current start so a just-completed
 * recurring task (whose start the backend has already advanced) does not draw
 * both the finished slot and its next occurrence.
 */
export function expandOccurrences(task: ITask, from: Date, to: Date): PlannedOccurrence[] {
	if (!task.startDate || !task.endDate) {
		return []
	}

	const realStart = dayjs(task.startDate)
	const realEnd = dayjs(task.endDate)
	const durationMs = realEnd.diff(realStart)
	const rangeStart = dayjs(from)
	const rangeEnd = dayjs(to)

	const occurrences: PlannedOccurrence[] = []
	const pushIfVisible = (start: dayjs.Dayjs, isGhost: boolean, index: number) => {
		const end = start.add(durationMs, 'millisecond')
		if (end.isAfter(rangeStart) && start.isBefore(rangeEnd)) {
			occurrences.push({
				key: `${task.id}-${index}`,
				task,
				start: start.toDate(),
				end: end.toDate(),
				isGhost,
			})
		}
	}

	pushIfVisible(realStart, false, 0)

	const step = getRepeatStep(task)
	if (step === null) {
		return occurrences
	}

	let cursor = realStart
	for (let i = 1; i <= MAX_OCCURRENCES; i++) {
		cursor = cursor.add(step.amount, step.unit)
		if (cursor.isAfter(rangeEnd)) {
			break
		}
		pushIfVisible(cursor, true, i)
	}

	return occurrences
}
