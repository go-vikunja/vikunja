import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {taskQuery, type TaskExpansion} from '@/client/queries/tasks'

export function useTask(id: MaybeRefOrGetter<number>, expand: MaybeRefOrGetter<TaskExpansion> = []) {
	const query = useQuery(computed(() => ({
		...taskQuery(toValue(id), toValue(expand)),
		// Both readers render the failure themselves: the detail view routes a missing task to
		// not-found, a task link renders an inline pill.
		meta: {handlesError: true},
	})))
	return {...query, task: query.data}
}
