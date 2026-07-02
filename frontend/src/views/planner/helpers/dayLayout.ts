import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import {expandOccurrences, allDayOccurrenceForDay} from './expandOccurrences'
import {packColumns} from './packColumns'
import type {PlannedOccurrence} from './types'

export interface AllDayItem {
	task: ITask
	isGhost: boolean
}

export interface TimedBlock {
	occurrence: PlannedOccurrence
	col: number
	cols: number
	topMinutes: number
	durationMinutes: number
}

const MIN_BLOCK_MINUTES = 15

// A task with a start and end both pinned to local midnight has no time-of-day
// and belongs in the all-day row, not as a (zero-height or full-column) block.
export function isAllDayTask(task: ITask): boolean {
	if (!task.startDate || !task.endDate) {
		return false
	}
	const start = dayjs(task.startDate)
	const end = dayjs(task.endDate)
	return start.hour() === 0 && start.minute() === 0 && end.hour() === 0 && end.minute() === 0
}

export function timedBlocksForDay(tasks: ITask[], day: Date): TimedBlock[] {
	const dayStart = dayjs(day).startOf('day')
	const dayEnd = dayStart.add(1, 'day')

	const occurrences: PlannedOccurrence[] = []
	for (const task of tasks) {
		if (!task.startDate || !task.endDate || isAllDayTask(task)) {
			continue
		}
		occurrences.push(...expandOccurrences(task, dayStart.toDate(), dayEnd.toDate()))
	}

	const sized = occurrences.map(occurrence => {
		const start = dayjs(occurrence.start).isBefore(dayStart) ? dayStart : dayjs(occurrence.start)
		const end = dayjs(occurrence.end).isAfter(dayEnd) ? dayEnd : dayjs(occurrence.end)
		return {
			occurrence,
			topMinutes: start.diff(dayStart, 'minute'),
			durationMinutes: Math.max(end.diff(start, 'minute'), MIN_BLOCK_MINUTES),
		}
	})

	return packColumns(
		sized,
		s => s.topMinutes,
		s => s.topMinutes + s.durationMinutes,
	).map(packed => ({...packed.item, col: packed.col, cols: packed.cols}))
}

export function allDayTasksForDay(tasks: ITask[], day: Date): AllDayItem[] {
	const target = dayjs(day)
	const items: AllDayItem[] = []
	for (const task of tasks) {
		if (isAllDayTask(task)) {
			const {covered, isGhost} = allDayOccurrenceForDay(task, day)
			if (covered) {
				items.push({task, isGhost})
			}
		} else if (!task.startDate && !task.endDate && task.dueDate && target.isSame(dayjs(task.dueDate), 'day')) {
			// due-only tasks (no time block) show on their due day
			items.push({task, isGhost: false})
		}
	}
	return items
}
