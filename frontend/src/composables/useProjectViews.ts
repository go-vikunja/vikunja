import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {projectQuery} from '@/client/queries/projects'

// Views come embedded in every project response, so this reads the project query
// instead of a separate endpoint - and stays live as mutations patch that cache.
export function useProjectViews(projectId: MaybeRefOrGetter<number>) {
	const query = useQuery(computed(() => projectQuery(Number(toValue(projectId)))))
	return {
		views: computed(() => query.data.value?.views ?? []),
		isPending: query.isPending,
		error: query.error,
	}
}
