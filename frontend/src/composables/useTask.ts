import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {taskQuery, type TaskExpansion} from '@/client/queries/tasks'

export function useTask(id: MaybeRefOrGetter<number>, expand: MaybeRefOrGetter<TaskExpansion> = []) {
	const query = useQuery(computed(() => taskQuery(toValue(id), toValue(expand))))
	return {...query, task: query.data}
}
