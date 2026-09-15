import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {storeToRefs} from 'pinia'

import {queryClient} from '@/client/queryClient'
import {projectQuery, projectsQuery} from '@/client/queries/projects'
import {useBaseStore} from '@/stores/base'

export function useCurrentProject() {
	const {currentProjectId} = storeToRefs(useBaseStore())
	const detailQuery = useQuery(computed(() => ({
		...projectQuery(currentProjectId.value),
		enabled: currentProjectId.value > 0,
	})), queryClient)
	const navigationQuery = useQuery(computed(() => ({
		...projectsQuery(),
		enabled: currentProjectId.value < 0,
	})), queryClient)

	return {
		currentProject: computed(() => {
			if (currentProjectId.value > 0) {
				return detailQuery.data.value ?? null
			}
			if (currentProjectId.value === -1) {
				return navigationQuery.data.value?.favoriteProject ?? null
			}
			return navigationQuery.data.value?.savedFilterProjects.find(
				project => project.id === currentProjectId.value,
			) ?? null
		}),
		isPending: computed(() => currentProjectId.value > 0
			? detailQuery.isPending.value
			: navigationQuery.isPending.value),
	}
}
