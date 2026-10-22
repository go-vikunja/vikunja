import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {addDays, isSameDay, startOfDay} from '@/helpers/time/dateMath'

export const DATE_SHORTCUT_KEYS = [
	'today',
	'tomorrow',
	'nextMonday',
	'thisWeekend',
	'laterThisWeek',
	'nextWeek',
	'nextWeekend',
	'endOfMonth',
	'inAMonth',
] as const
export type DateShortcutKey = typeof DATE_SHORTCUT_KEYS[number]

function resolve(key: DateShortcutKey, now: Date): Date {
	switch (key) {
		case 'nextWeekend': {
			const daysUntilSaturday = (6 - now.getDay() + 7) % 7
			// On Sunday the weekend just ended, so the very next Saturday already is "next weekend" — no extra week to skip.
			return addDays(now, now.getDay() === 0 ? daysUntilSaturday : daysUntilSaturday + 7)
		}
		case 'endOfMonth':
			return new Date(now.getFullYear(), now.getMonth() + 1, 0)
		case 'inAMonth': {
			const date = new Date(now)
			const lastDay = new Date(now.getFullYear(), now.getMonth() + 2, 0).getDate()
			date.setMonth(date.getMonth() + 1, Math.min(now.getDate(), lastDay))
			return date
		}
		default:
			return addDays(now, calculateDayInterval(key, now.getDay()))
	}
}

export interface DateShortcut {
	key: DateShortcutKey
	date: Date
}

// Late in the day "today" is pointless, and on a late Sunday "this weekend" would be today as well.
function isPointlessLate(key: DateShortcutKey, now: Date): boolean {
	if (now.getHours() < 21) {
		return false
	}
	return key === 'today' || (key === 'thisWeekend' && now.getDay() === 0)
}

// Several shortcuts can land on the same day (Friday: tomorrow = this weekend); the earlier, more basic one wins so no two entries highlight together.
export function buildDateShortcuts(now: Date = new Date()): DateShortcut[] {
	const shortcuts: DateShortcut[] = []
	for (const key of DATE_SHORTCUT_KEYS) {
		if (isPointlessLate(key, now)) {
			continue
		}
		const date = startOfDay(resolve(key, now))
		if (shortcuts.some(existing => isSameDay(existing.date, date))) {
			continue
		}
		shortcuts.push({key, date})
	}
	return shortcuts
}
