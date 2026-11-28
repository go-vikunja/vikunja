import dayjs from 'dayjs'
import isoWeek from 'dayjs/plugin/isoWeek'

dayjs.extend(isoWeek)

// Zoom is the narrowest a day may get, in pixels. The chart widens days to fill its
// container when the range is short, so the minimum zoom means "fit the range".
export const GANTT_ZOOM_MIN = 2
export const GANTT_ZOOM_MAX = 120
export const DEFAULT_GANTT_ZOOM = 30
export const GANTT_ZOOM_STEP = 1.5

export const GANTT_ZOOM_PRESETS = {
	day: DEFAULT_GANTT_ZOOM,
	week: 8,
	month: 3,
} as const

export function clampGanttZoom(zoom: number): number {
	const clamped = Math.min(GANTT_ZOOM_MAX, Math.max(GANTT_ZOOM_MIN, zoom))
	return Math.round(clamped * 100) / 100
}

export function parseGanttZoom(value: unknown): number | undefined {
	if (typeof value !== 'string' && typeof value !== 'number') {
		return undefined
	}
	const zoom = Number(value)
	return Number.isFinite(zoom) && zoom > 0
		? clampGanttZoom(zoom)
		: undefined
}

export function scaleGanttZoom(zoom: number, factor: number): number {
	return clampGanttZoom(zoom * factor)
}

export function stepGanttZoom(zoom: number, direction: 1 | -1): number {
	return scaleGanttZoom(zoom, Math.pow(GANTT_ZOOM_STEP, direction))
}

export type GanttTimelineUnit = 'day' | 'week' | 'month' | 'year'

export interface GanttTimelineUnits {
	upper: GanttTimelineUnit
	lower: GanttTimelineUnit
	showWeekday: boolean
}

const DAY_LABEL_MIN_WIDTH = 24
const WEEKDAY_LABEL_MIN_WIDTH = 30
const WEEK_LABEL_MIN_WIDTH = 36

export function timelineUnitsForDayWidth(dayWidthPixels: number): GanttTimelineUnits {
	if (dayWidthPixels >= DAY_LABEL_MIN_WIDTH) {
		return {upper: 'month', lower: 'day', showWeekday: dayWidthPixels >= WEEKDAY_LABEL_MIN_WIDTH}
	}
	if (dayWidthPixels * 7 >= WEEK_LABEL_MIN_WIDTH) {
		return {upper: 'month', lower: 'week', showWeekday: false}
	}
	return {upper: 'year', lower: 'month', showWeekday: false}
}

export interface GanttTimelineGroup {
	key: string
	start: Date
	dayCount: number
}

export function isoWeekNumber(date: Date): number {
	return dayjs(date).isoWeek()
}

function timelineUnitKey(date: Date, unit: GanttTimelineUnit): string {
	const day = dayjs(date)
	switch (unit) {
		case 'day':
			return day.format('YYYY-MM-DD')
		case 'week':
			return `${day.isoWeekYear()}-W${day.isoWeek()}`
		case 'month':
			return day.format('YYYY-MM')
		case 'year':
			return day.format('YYYY')
	}
}

export function groupTimelineDates(dates: Date[], unit: GanttTimelineUnit): GanttTimelineGroup[] {
	const groups: GanttTimelineGroup[] = []

	for (const date of dates) {
		const key = timelineUnitKey(date, unit)
		const lastGroup = groups[groups.length - 1]

		if (lastGroup?.key === key) {
			lastGroup.dayCount++
		} else {
			groups.push({key, start: date, dayCount: 1})
		}
	}

	return groups
}

export function timelineGroupOffsets(groups: GanttTimelineGroup[], dayWidthPixels: number): number[] {
	const offsets: number[] = []
	let days = 0

	for (const group of groups) {
		offsets.push(days * dayWidthPixels)
		days += group.dayCount
	}

	return offsets
}
