import type {ITask} from '@/modelTypes/ITask'

// A single concrete placement of a task on the calendar grid. Recurring tasks
// expand into one real occurrence (the stored task) plus dimmed, read-only
// ghost occurrences projected forward across the visible range.
export interface PlannedOccurrence {
	key: string
	task: ITask
	start: Date
	end: Date
	isGhost: boolean
}
