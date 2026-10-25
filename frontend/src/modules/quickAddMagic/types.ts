import type {RepeatFrequency} from '@/helpers/rrule'

// Quick add only recognises the units the simple repeat editor offers.
export interface QuickAddRepeat {
	type: RepeatFrequency,
	amount: number,
}

export interface repeatParsedResult {
	textWithoutMatched: string,
	repeats: QuickAddRepeat | null,
}

export interface ParsedTaskText {
	text: string,
	date: Date | null,
	labels: string[],
	project: string | null,
	priority: number | null,
	assignees: string[],
	repeats: QuickAddRepeat | null,
}

export interface Prefixes {
	label: string,
	project: string,
	priority: string,
	assignee: string,
}
