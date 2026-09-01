import type {TaskCollection} from '@/client/generated'

export type ExpandTaskFilterParam = 'subtasks' | 'buckets' | 'reactions' | 'comment_count' | 'is_unread' | null

export interface TaskFilterParams {
	sort_by: ('start_date' | 'end_date' | 'due_date' | 'done' | 'id' | 'position' | 'title' | 'relevance')[],
	order_by: ('asc' | 'desc')[],
	filter: string,
	filter_include_nulls: boolean,
	filter_timezone?: string,
	s: string,
	per_page?: number,
	expand?: ExpandTaskFilterParam,
}

export type EditableTaskCollection = Required<Omit<TaskCollection, 'sort_by' | 'order_by'>> &
	Pick<TaskFilterParams, 'sort_by' | 'order_by'>
