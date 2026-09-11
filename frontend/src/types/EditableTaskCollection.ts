import type {TaskCollection} from '@/client/generated'

export type EditableTaskCollection = Required<Omit<TaskCollection, 'sort_by' | 'order_by'>> & {
	sort_by: string[]
	order_by: string[]
}
