import {MILLISECONDS_A_DAY} from '@/constants/date'

// 'fit' keeps the original behaviour: the visible range is stretched to the width of the chart.
export type GanttZoom = 'fit' | 'day' | 'week' | 'month' | 'quarter'

export const GANTT_ZOOMS: GanttZoom[] = ['fit', 'day', 'week', 'month', 'quarter']

// Pixels per day at each fixed zoom level.
export const ZOOM_DAY_WIDTH: Record<Exclude<GanttZoom, 'fit'>, number> = {
	day: 40,
	week: 14,
	month: 5,
	quarter: 2,
}

const FIT_MIN_DAY_WIDTH = 30

export function isGanttZoom(value: unknown): value is GanttZoom {
	return typeof value === 'string' && (GANTT_ZOOMS as string[]).includes(value)
}

export function resolveDayWidth(zoom: GanttZoom, fitWidth: number): number {
	if (zoom === 'fit') {
		return Math.max(fitWidth, FIT_MIN_DAY_WIDTH)
	}
	return ZOOM_DAY_WIDTH[zoom]
}

export type TierUnit = 'year' | 'quarter' | 'month' | 'week'

export interface TimelineCell {
	key: string
	unit: TierUnit
	start: Date
	x: number
	width: number
	label: string
}

export interface TimelineTiers {
	upper: TimelineCell[]
	// Empty for day and fit: the header draws one cell per day itself.
	lower: TimelineCell[]
}

export type TierLabelFormatter = (date: Date, unit: TierUnit) => string

const defaultFormat: TierLabelFormatter = (date, unit) => {
	const y = date.getFullYear()
	const m = String(date.getMonth() + 1).padStart(2, '0')
	switch (unit) {
		case 'year': return String(y)
		case 'quarter': return `${y} Q${Math.floor(date.getMonth() / 3) + 1}`
		case 'month': return `${y}-${m}`
		case 'week': return `${y}-${m}-${String(date.getDate()).padStart(2, '0')}`
	}
}

// Monday of the week of a local date.
export function startOfWeek(date: Date): Date {
	const d = new Date(date.getFullYear(), date.getMonth(), date.getDate())
	const offset = (d.getDay() + 6) % 7
	d.setDate(d.getDate() - offset)
	return d
}

function unitKey(date: Date, unit: TierUnit): string {
	const y = date.getFullYear()
	switch (unit) {
		case 'year': return String(y)
		case 'quarter': return `${y}-Q${Math.floor(date.getMonth() / 3)}`
		case 'month': return `${y}-${date.getMonth()}`
		case 'week': {
			const monday = startOfWeek(date)
			return `${monday.getFullYear()}-${monday.getMonth()}-${monday.getDate()}`
		}
	}
}

function groupCells(dates: Date[], dayWidth: number, unit: TierUnit, format: TierLabelFormatter): TimelineCell[] {
	const cells: TimelineCell[] = []
	dates.forEach((date, index) => {
		const key = unitKey(date, unit)
		const last = cells[cells.length - 1]
		if (last && last.key === key) {
			last.width += dayWidth
			return
		}
		cells.push({key, unit, start: date, x: index * dayWidth, width: dayWidth, label: format(date, unit)})
	})
	return cells
}

const TIERS: Record<GanttZoom, {upper: TierUnit, lower: TierUnit | null}> = {
	fit: {upper: 'month', lower: null},
	day: {upper: 'month', lower: null},
	week: {upper: 'month', lower: 'week'},
	month: {upper: 'year', lower: 'month'},
	quarter: {upper: 'year', lower: 'quarter'},
}

export function buildTimelineTiers(
	dates: Date[],
	dayWidth: number,
	zoom: GanttZoom,
	format: TierLabelFormatter = defaultFormat,
): TimelineTiers {
	const {upper, lower} = TIERS[zoom]
	return {
		upper: groupCells(dates, dayWidth, upper, format),
		lower: lower ? groupCells(dates, dayWidth, lower, format) : [],
	}
}

// X positions of the vertical grid lines: every day while days are readable, otherwise the
// boundaries of the lower header tier.
export function gridLineXs(dates: Date[], dayWidth: number, zoom: GanttZoom): number[] {
	if (zoom === 'fit' || zoom === 'day') {
		return dates.map((_, index) => index * dayWidth)
	}
	return buildTimelineTiers(dates, dayWidth, zoom).lower.map(cell => cell.x)
}

export interface WeekendBand {
	x: number
	width: number
}

// Saturday and Sunday are shaded while a day is wide enough to see (not at month and quarter zoom).
export function weekendBands(dates: Date[], dayWidth: number, zoom: GanttZoom): WeekendBand[] {
	if (zoom === 'month' || zoom === 'quarter') {
		return []
	}
	const bands: WeekendBand[] = []
	dates.forEach((date, index) => {
		const day = date.getDay()
		if (day !== 0 && day !== 6) {
			return
		}
		const x = index * dayWidth
		const last = bands[bands.length - 1]
		if (last && Math.abs(last.x + last.width - x) < 0.001) {
			last.width += dayWidth
			return
		}
		bands.push({x, width: dayWidth})
	})
	return bands
}

// X position of "now", or null when it is outside the shown days.
export function todayX(dates: Date[], dayWidth: number, now: Date): number | null {
	const first = dates[0]
	if (!first) {
		return null
	}
	const dayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
	const firstStart = new Date(first.getFullYear(), first.getMonth(), first.getDate())
	const index = Math.round((dayStart.getTime() - firstStart.getTime()) / MILLISECONDS_A_DAY)
	if (index < 0 || index >= dates.length) {
		return null
	}
	const fraction = (now.getTime() - dayStart.getTime()) / MILLISECONDS_A_DAY
	return index * dayWidth + fraction * dayWidth
}
