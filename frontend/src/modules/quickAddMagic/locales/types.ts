import type {IRepeatType} from '@/types/IRepeatAfter'

export interface WeekdayDef {
	/** Inflected forms and abbreviations, longest first */
	forms: string[],
	/** JS Date.getDay() index: 0 = Sunday */
	day: number,
}

export interface MonthDef {
	/** Nominative and genitive forms ("серпень"/"серпня") */
	forms: string[],
	/** JS Date month index: 0 = January */
	month: number,
}

export interface RepeatUnitDef {
	forms: string[],
	type: IRepeatType,
}

export interface RepeatAdverbDef {
	forms: string[],
	type: IRepeatType,
	amount: number,
}

export interface NumberWordDef {
	forms: string[],
	value: number,
}

export type DatePhrase =
	'today' | 'tonight' | 'tomorrow' | 'nextMonday' |
	'thisWeekend' | 'laterThisWeek' | 'laterNextWeek' |
	'nextWeek' | 'nextMonth' | 'endOfMonth'

export interface QuickAddMagicLocale {
	code: string,
	/** Fixed date phrases. Evaluated in object insertion order — list longer phrases
	 * before their prefixes (e.g. "сьогодні ввечері" before "сьогодні"). */
	phrases: Partial<Record<DatePhrase, string[]>>,
	weekdays: WeekdayDef[],
	/** Words allowed directly before a weekday ("next monday", "в понеділок") */
	weekdayPrefixes: string[],
	months: MonthDef[],
	/** Suffixes accepted after a day number ("21st", "21-го") */
	ordinalSuffixes: string[],
	/** Words introducing a time of day ("at 14:00", "о 14:00") */
	timePrefixes: string[],
	/** Words introducing a relative offset ("in 3 days", "через 3 дні") */
	inPrefixes: string[],
	everyKeywords: string[],
	numberWords: NumberWordDef[],
	repeatUnits: RepeatUnitDef[],
	repeatAdverbs: RepeatAdverbDef[],
	/** \b works for Latin scripts only; Cyrillic needs space boundaries instead */
	monthWordBoundary?: boolean,
	/** Require a space/end directly after a date phrase so words merely starting
	 * with a phrase don't match ("завтрак" ≠ "завтра"). Off keeps legacy behavior. */
	phraseStrictBoundary?: boolean,
}
