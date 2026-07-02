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

// Skip ahead to shortly before `towards` so a task whose stored start is far in
// the past (e.g. a daily repeater untouched for years) doesn't exhaust the
// iteration cap before reaching the visible range. `minBackoffMs` keeps enough
// margin that an occurrence starting before the window but still overlapping it
// isn't skipped. The caller's fine-stepping loop covers the small remainder.
function coarseJump(
	realStart: dayjs.Dayjs,
	step: {amount: number, unit: ManipulateType},
	towards: dayjs.Dayjs,
	minBackoffMs: number,
): {cursor: dayjs.Dayjs, index: number} {
	const stepMs = realStart.add(step.amount, step.unit).diff(realStart)
	if (stepMs <= 0 || !realStart.isBefore(towards)) {
		return {cursor: realStart, index: 0}
	}
	const backoffSteps = Math.ceil(minBackoffMs / stepMs) + 1
	const jumps = Math.max(Math.floor(towards.diff(realStart) / stepMs) - backoffSteps, 0)
	return {cursor: realStart.add(step.amount * jumps, step.unit), index: jumps}
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

	let {cursor, index} = coarseJump(realStart, step, rangeStart, durationMs)
	for (let i = 0; i < MAX_OCCURRENCES; i++) {
		cursor = cursor.add(step.amount, step.unit)
		index++
		if (cursor.isAfter(rangeEnd)) {
			break
		}
		pushIfVisible(cursor, true, index)
	}

	return occurrences
}

/**
 * Whether an all-day task covers `day`, following its recurrence. All-day
 * occurrences sit at midnight with zero duration, which expandOccurrences'
 * range test excludes, so they get their own day-granular check here.
 * Returns isGhost = true when only a projected occurrence (not the stored
 * span) lands on the day, so the caller can render it read-only.
 */
export function allDayOccurrenceForDay(task: ITask, day: Date): {covered: boolean, isGhost: boolean} {
	if (!task.startDate || !task.endDate) {
		return {covered: false, isGhost: false}
	}

	const target = dayjs(day).startOf('day')
	const realStart = dayjs(task.startDate).startOf('day')
	const realEnd = dayjs(task.endDate).startOf('day')
	const spanDays = Math.max(realEnd.diff(realStart, 'day'), 0)
	const covers = (start: dayjs.Dayjs) => !target.isBefore(start) && !target.isAfter(start.add(spanDays, 'day'))

	if (covers(realStart)) {
		return {covered: true, isGhost: false}
	}

	const step = getRepeatStep(task)
	if (step === null) {
		return {covered: false, isGhost: false}
	}

	// Back off by the task's span so a long all-day occurrence starting before
	// the target day but still covering it isn't jumped over.
	let {cursor} = coarseJump(realStart, step, target, spanDays * 24 * 60 * 60 * 1000)
	for (let i = 0; i < MAX_OCCURRENCES; i++) {
		cursor = cursor.add(step.amount, step.unit)
		if (cursor.isAfter(target)) {
			break
		}
		if (covers(cursor)) {
			return {covered: true, isGhost: true}
		}
	}

	return {covered: false, isGhost: false}
}
