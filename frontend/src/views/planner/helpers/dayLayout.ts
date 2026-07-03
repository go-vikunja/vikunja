import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import {expandOccurrences, allDayCoveredDays} from './expandOccurrences'
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

export function dayKey(day: Date | dayjs.Dayjs): string {
	return dayjs(day).format('YYYY-MM-DD')
}

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

// Expand every task once across the whole visible range and bucket the
// occurrences into the days they overlap — one recurrence walk per task
// instead of one per (task, day).
export function timedBlocksByDay(tasks: ITask[], days: Date[]): Map<string, TimedBlock[]> {
	const blocks = new Map<string, TimedBlock[]>()
	if (days.length === 0) {
		return blocks
	}

	const sizedByDay = new Map<string, Array<{occurrence: PlannedOccurrence, topMinutes: number, durationMinutes: number}>>()
	for (const day of days) {
		sizedByDay.set(dayKey(day), [])
	}
	const rangeStart = dayjs(days[0]).startOf('day')
	const rangeEnd = dayjs(days[days.length - 1]).startOf('day').add(1, 'day')

	for (const task of tasks) {
		if (!task.startDate || !task.endDate || isAllDayTask(task)) {
			continue
		}
		for (const occurrence of expandOccurrences(task, rangeStart.toDate(), rangeEnd.toDate())) {
			const occStart = dayjs(occurrence.start)
			const occEnd = dayjs(occurrence.end)

			if (!occEnd.isAfter(occStart)) {
				// Degenerate zero-length occurrence: render a minimum-height block
				// on its day instead of dropping it.
				const day = occStart.startOf('day')
				sizedByDay.get(dayKey(day))?.push({
					occurrence,
					topMinutes: occStart.diff(day, 'minute'),
					durationMinutes: MIN_BLOCK_MINUTES,
				})
				continue
			}

			let day = occStart.startOf('day')
			if (day.isBefore(rangeStart)) {
				day = rangeStart
			}
			for (; day.isBefore(rangeEnd) && day.isBefore(occEnd); day = day.add(1, 'day')) {
				const dayEnd = day.add(1, 'day')
				const start = occStart.isBefore(day) ? day : occStart
				const end = occEnd.isAfter(dayEnd) ? dayEnd : occEnd
				sizedByDay.get(dayKey(day))?.push({
					occurrence,
					topMinutes: start.diff(day, 'minute'),
					durationMinutes: Math.max(end.diff(start, 'minute'), MIN_BLOCK_MINUTES),
				})
			}
		}
	}

	for (const [key, sized] of sizedByDay) {
		blocks.set(key, packColumns(
			sized,
			s => s.topMinutes,
			s => s.topMinutes + s.durationMinutes,
		).map(packed => ({...packed.item, col: packed.col, cols: packed.cols})))
	}
	return blocks
}

export function allDayItemsByDay(tasks: ITask[], days: Date[]): Map<string, AllDayItem[]> {
	const map = new Map<string, AllDayItem[]>()
	if (days.length === 0) {
		return map
	}
	for (const day of days) {
		map.set(dayKey(day), [])
	}
	const from = days[0]
	const to = dayjs(days[days.length - 1]).add(1, 'day').toDate()

	for (const task of tasks) {
		if (isAllDayTask(task)) {
			for (const [key, isGhost] of allDayCoveredDays(task, from, to)) {
				map.get(key)?.push({task, isGhost})
			}
		} else if (!task.startDate || !task.endDate) {
			// Tasks with no time block (due-only, start-only or end-only) anchor
			// to their single date so they don't vanish from the planner.
			const anchor = task.startDate ?? task.endDate ?? task.dueDate
			if (anchor) {
				map.get(dayKey(anchor))?.push({task, isGhost: false})
			}
		}
	}
	return map
}
