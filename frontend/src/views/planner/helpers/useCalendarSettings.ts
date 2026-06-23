import {useStorage} from '@vueuse/core'

export interface CalendarSettings {
	// Working hours ("HH:MM") define the initial zoom/scroll window — the grid
	// still renders the full 0–24h so off-hours stay reachable by scrolling.
	dayStart: string
	dayEnd: string
	defaultDurationMinutes: number
	slotMinutes: number
	showDone: boolean
	// true: week aligned to the user's first weekday; false: 7 days from today.
	fullWeek: boolean
}

const DEFAULTS: CalendarSettings = {
	dayStart: '08:00',
	dayEnd: '18:00',
	defaultDurationMinutes: 60,
	slotMinutes: 30,
	showDone: false,
	fullWeek: true,
}

// Module-level so every caller shares the same reactive ref within the tab.
const settings = useStorage<CalendarSettings>('planner-settings', {...DEFAULTS}, localStorage, {mergeDefaults: true})

export function useCalendarSettings() {
	return {settings}
}
