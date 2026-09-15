import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {projectViewsQuery} from '@/client/queries/projectViews'

export function useProjectViews(projectId: MaybeRefOrGetter<number>) {
	const query = useQuery(computed(() => projectViewsQuery(toValue(projectId))))
	return {
		views: computed(() => query.data.value ?? []),
		isPending: query.isPending,
		error: query.error,
	}
}
