import dayjs from 'dayjs'

import type {ITask} from '@/modelTypes/ITask'
import type {IRepeatAfter} from '@/types/IRepeatAfter'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import type {PlannedOccurrence} from './types'

// Guard against pathological repeat intervals projecting forever.
const MAX_OCCURRENCES = 366

const TYPE_TO_SECONDS: Record<IRepeatAfter['type'], number> = {
	seconds: 1,
	minutes: 60,
	hours: 3600,
	days: 86400,
	weeks: 604800,
	// The task edit UI only produces hours/days/weeks; these are defensive.
	months: 30 * 86400,
	years: 365 * 86400,
}

type RepeatStep =
	| {kind: 'fixed', seconds: number}
	| {kind: 'month'}

// Mirrors the backend's addOneMonthToDate (pkg/models/tasks.go): same day next
// month, overflowing like Go's time.Date instead of clamping — Jan 31 becomes
// Mar 2/3, not Feb 28 — so ghosts land on the days the backend will schedule.
function addOneMonthWithOverflow(date: dayjs.Dayjs): dayjs.Dayjs {
	return date.date(1).add(1, 'month').add(date.date() - 1, 'day')
}

function advance(date: dayjs.Dayjs, step: RepeatStep): dayjs.Dayjs {
	return step.kind === 'month'
		? addOneMonthWithOverflow(date)
		// The backend adds a fixed number of seconds (not calendar units), so
		// wall times shift across DST exactly like they will on the server.
		: date.add(step.seconds, 'second')
}

function getRepeatStep(task: ITask): RepeatStep | null {
	// Monthly mode repeats on the same day each month regardless of repeatAfter.
	if (task.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH) {
		return {kind: 'month'}
	}

	// From-current-date mode computes the next occurrence from the moment the
	// task is completed, which is unknowable ahead of time — projected ghosts
	// would show slots the backend will never schedule, so don't project any.
	if (task.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_FROM_CURRENT_DATE) {
		return null
	}

	const seconds = typeof task.repeatAfter === 'number'
		? task.repeatAfter
		: (task.repeatAfter?.amount ?? 0) * TYPE_TO_SECONDS[task.repeatAfter?.type ?? 'seconds']

	if (seconds <= 0) {
		return null
	}

	return {kind: 'fixed', seconds}
}

// Skip ahead to shortly before `towards` so a task whose stored start is far in
// the past (e.g. a daily repeater untouched for years) doesn't exhaust the
// iteration cap before reaching the visible range. `minBackoffMs` keeps enough
// margin that an occurrence starting before the window but still overlapping it
// isn't skipped. The caller's fine-stepping loop covers the small remainder.
function coarseJump(
	realStart: dayjs.Dayjs,
	step: RepeatStep,
	towards: dayjs.Dayjs,
	minBackoffMs: number,
): {cursor: dayjs.Dayjs, index: number} {
	// Month steps depend on the previous occurrence (overflow can change the
	// day-of-month), so they can't be jumped multiplicatively. The iteration
	// cap still covers ~30 years of monthly steps.
	if (step.kind === 'month') {
		return {cursor: realStart, index: 0}
	}
	const stepMs = step.seconds * 1000
	if (!realStart.isBefore(towards)) {
		return {cursor: realStart, index: 0}
	}
	const jumps = Math.max(Math.floor((towards.diff(realStart) - minBackoffMs) / stepMs) - 1, 0)
	return {cursor: realStart.add(jumps * step.seconds, 'second'), index: jumps}
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
		cursor = advance(cursor, step)
		index++
		if (cursor.isAfter(rangeEnd)) {
			break
		}
		pushIfVisible(cursor, true, index)
	}

	return occurrences
}

/**
 * Day keys ('YYYY-MM-DD') an all-day task covers within [from, to), following
 * its recurrence, mapped to whether the coverage is only a projected (ghost)
 * occurrence. All-day occurrences sit at midnight with zero duration, which
 * expandOccurrences' range test excludes, so they get their own day-granular
 * expansion here — one recurrence walk per task for the whole range.
 */
export function allDayCoveredDays(task: ITask, from: Date, to: Date): Map<string, boolean> {
	const covered = new Map<string, boolean>()
	if (!task.startDate || !task.endDate) {
		return covered
	}

	const rangeStart = dayjs(from).startOf('day')
	const rangeEnd = dayjs(to).startOf('day')
	const realStart = dayjs(task.startDate).startOf('day')
	const realEnd = dayjs(task.endDate).startOf('day')
	const spanDays = Math.max(realEnd.diff(realStart, 'day'), 0)

	const mark = (start: dayjs.Dayjs, isGhost: boolean) => {
		let day = start.startOf('day')
		const last = day.add(spanDays, 'day')
		if (day.isBefore(rangeStart)) {
			day = rangeStart
		}
		for (; !day.isAfter(last) && day.isBefore(rangeEnd); day = day.add(1, 'day')) {
			const key = day.format('YYYY-MM-DD')
			// Real coverage wins over ghost coverage on the same day.
			if (!isGhost || !covered.has(key)) {
				covered.set(key, isGhost)
			}
		}
	}

	mark(realStart, false)

	const step = getRepeatStep(task)
	if (step === null) {
		return covered
	}

	// Back off by the task's span so a long all-day occurrence starting before
	// the range but still reaching into it isn't jumped over.
	let {cursor} = coarseJump(realStart, step, rangeStart, (spanDays + 1) * 24 * 60 * 60 * 1000)
	for (let i = 0; i < MAX_OCCURRENCES; i++) {
		cursor = advance(cursor, step)
		if (cursor.startOf('day').isAfter(rangeEnd)) {
			break
		}
		mark(cursor, true)
	}

	return covered
}
