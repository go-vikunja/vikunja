import {computed} from 'vue'
import {useQuery} from '@tanstack/vue-query'

import {
	findProjectByExactTitle,
	findProjectByIdentifier,
	getProjectById,
	projectsQuery,
	searchProjects,
} from '@/client/queries/projects'

export function useProjects() {
	const query = useQuery(projectsQuery())
	const projects = computed(() => query.data.value?.projects ?? [])
	const projectMap = computed(() => Object.fromEntries(
		projects.value.map(project => [project.id, project]),
	))

	return {
		projects,
		projectMap,
		isPending: query.isPending,
		getProjectById: (id: number) => getProjectById(projects.value, id),
		findProjectByExactTitle: (title: string) => findProjectByExactTitle(projects.value, title),
		findProjectByIdentifier: (identifier: string) => findProjectByIdentifier(projects.value, identifier),
		searchProjects: (value: string, includeArchived = false) =>
			searchProjects(projects.value, value, includeArchived),
	}
}
