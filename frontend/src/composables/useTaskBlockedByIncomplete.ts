import {computed, type Ref} from 'vue'

import type {ITask} from '@/modelTypes/ITask'
import {RELATION_KIND} from '@/types/IRelationKind'

export function useTaskBlockedByIncomplete(task: Ref<ITask>) {
	return computed(() =>
		task.value.relatedTasks?.[RELATION_KIND.BLOCKED]?.some(relatedTask => !relatedTask.done) ?? false,
	)
}
