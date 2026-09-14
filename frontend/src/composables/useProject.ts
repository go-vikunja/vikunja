import {computed, ref, toValue, watch, type MaybeRefOrGetter} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {
	normalizeProject,
	projectQuery,
} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'

export function useProject(projectId: MaybeRefOrGetter<number>) {
	const query = useQuery(computed(() => projectQuery(Number(toValue(projectId)))))
	const project = ref<ProjectResponse>(normalizeProject({id: Number(toValue(projectId))}))
	const loadedProjectId = ref(0)

	watch([
		() => Number(toValue(projectId)),
		query.data,
		query.isFetching,
		query.isError,
	], ([id, value, isFetching]) => {
		if (loadedProjectId.value !== id) {
			project.value = normalizeProject({id})
		}
		if (!isFetching && value && loadedProjectId.value !== id) {
			project.value = normalizeProject(value)
			loadedProjectId.value = id
		}
	}, {immediate: true})

	return {
		project,
		isLoading: computed(() =>
			query.isFetching.value ||
			(!query.isError.value && loadedProjectId.value !== Number(toValue(projectId))),
		),
		error: query.error,
		isLoaded: computed(() => loadedProjectId.value > 0 && loadedProjectId.value === Number(toValue(projectId))),
	}
}
