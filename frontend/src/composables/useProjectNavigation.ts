import {computed, reactive} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {createSharedComposable} from '@vueuse/core'

import {
	ensureProject,
	findProjectByExactTitle,
	findProjectByIdentifier,
	getChildProjects,
	getEffectiveParentProjectId,
	getFavoriteNavigationItems,
	getProjectAncestors,
	getRootProjects,
	projectsQuery,
	refreshProjects,
	searchProjects,
} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import {queryClient} from '@/client/queryClient'

// One observer and one derived list set for the dozens of consumers.
const useSharedProjectNavigation = createSharedComposable(() => {
	// The explicit client keeps this usable outside a Vue injection context.
	const query = useQuery(projectsQuery(), queryClient)
	const realProjects = computed(() => query.data.value?.projects ?? [])
	const favoriteProject = computed(() => query.data.value?.favoriteProject ?? null)
	const savedFilterProjects = computed(() => query.data.value?.savedFilterProjects ?? [])
	const projectsArray = computed(() => [
		...realProjects.value,
		...(favoriteProject.value ? [favoriteProject.value] : []),
		...savedFilterProjects.value,
	])
	const projects = computed(() => Object.fromEntries(
		projectsArray.value.map(project => [project.id, project]),
	))
	const rootProjects = computed(() => getRootProjects(realProjects.value))
	const favoriteProjects = computed(() => query.data.value
		? getFavoriteNavigationItems(query.data.value)
		: [])
	const hasProjects = computed(() => projectsArray.value.length > 0)

	return {
		query,
		realProjects,
		favoriteProject,
		savedFilterProjects,
		projectsArray,
		projects,
		rootProjects,
		favoriteProjects,
		hasProjects,
	}
})

export function useProjectNavigation() {
	const {
		query,
		realProjects,
		favoriteProject,
		savedFilterProjects,
		projectsArray,
		projects,
		rootProjects,
		favoriteProjects,
		hasProjects,
	} = useSharedProjectNavigation()

	const state = reactive({
		projects,
		projectsArray,
		notArchivedRootProjects: rootProjects,
		favoriteProject,
		savedFilterProjects,
		favoriteProjects,
		hasProjects,
		isLoading: query.isPending,
		getChildProjects: (id: number) => getChildProjects(realProjects.value, id),
		getAncestors: (project: ProjectResponse) => getProjectAncestors(realProjects.value, project),
		getEffectiveParentProjectId: (project: ProjectResponse, parentProjectIdFromDom: number) =>
			getEffectiveParentProjectId(realProjects.value, project, parentProjectIdFromDom),
		findProjectByExactname: (title: string) => findProjectByExactTitle(realProjects.value, title),
		findProjectByIdentifier: (identifier: string) => findProjectByIdentifier(realProjects.value, identifier),
		searchProject: (value: string, includeArchived = false) =>
			searchProjects(realProjects.value, value, includeArchived),
		searchSavedFilter: (value: string, includeArchived = false) => {
			const normalized = value.toLowerCase()
			return value === '' ? [] : savedFilterProjects.value.filter(project =>
				project.is_archived === includeArchived &&
				(project.title.toLowerCase().includes(normalized) ||
					project.description.toLowerCase().includes(normalized)),
			)
		},
		searchProjectAndFilter: (value: string, includeArchived = false) => {
			const normalized = value.toLowerCase()
			return value === '' ? [] : projectsArray.value.filter(project =>
				project.is_archived === includeArchived &&
				(project.title.toLowerCase().includes(normalized) ||
					project.description.toLowerCase().includes(normalized)),
			)
		},
		loadAllProjects: refreshProjects,
		loadProject: ensureProject,
	})

	return state
}
