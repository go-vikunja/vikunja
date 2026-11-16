import type {InjectionKey, Ref} from 'vue'

import type {TaskResponse} from '@/client/queries/tasks'

export const PLANNER_VIEW_MODES = ['month', 'week', 'day'] as const

export type PlannerViewMode = typeof PLANNER_VIEW_MODES[number]

// A single concrete placement of a task on the calendar grid. Recurring tasks
// expand into one real occurrence (the stored task) plus dimmed, read-only
// ghost occurrences projected forward across the visible range.
export interface PlannedOccurrence {
	key: string
	task: TaskResponse
	start: Date
	end: Date
	isGhost: boolean
}

// Day key of the all-day target (cell or day header) currently hovered by a
// drag. Provided by CalendarGrid; blocks update it during pointer drags so
// those get the same drop-target highlight as native HTML5 drags.
export const ALL_DAY_DROP_TARGET: InjectionKey<Ref<string | null>> = Symbol('plannerAllDayDropTarget')
