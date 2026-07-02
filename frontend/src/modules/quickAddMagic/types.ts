import type {ITaskRepeat} from '@/modelTypes/ITask'

export interface repeatParsedResult {
	textWithoutMatched: string,
	repeat: ITaskRepeat | null,
}

export interface ParsedTaskText {
	text: string,
	date: Date | null,
	labels: string[],
	project: string | null,
	priority: number | null,
	assignees: string[],
	repeat: ITaskRepeat | null,
}

export interface Prefixes {
	label: string,
	project: string,
	priority: string,
	assignee: string,
}
