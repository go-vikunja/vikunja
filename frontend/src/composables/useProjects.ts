import {computed, reactive} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {createSharedComposable} from '@vueuse/core'

import {
	ensureProject,
	findProjectByExactTitle,
	projectsQuery,
} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import {queryClient} from '@/client/queryClient'

function sortByPosition(projects: ProjectResponse[]): ProjectResponse[] {
	return [...projects].sort((a, b) => a.position - b.position)
}

function isOrphaned(projects: readonly ProjectResponse[], project: ProjectResponse): boolean {
	return project.parent_project_id !== 0 && !projects.some(parent => parent.id === project.parent_project_id)
}

function getAncestors(projects: readonly ProjectResponse[], project: ProjectResponse | undefined): ProjectResponse[] {
	if (!project) {
		return []
	}
	if (project.parent_project_id === 0) {
		return [project]
	}
	const parent = projects.find(candidate => candidate.id === project.parent_project_id)
	return [...getAncestors(projects, parent), project]
}

function search(projects: readonly ProjectResponse[], value: string, includeArchived: boolean): ProjectResponse[] {
	if (value === '') {
		return []
	}
	const normalized = value.toLowerCase()
	return projects.filter(project =>
		project.is_archived === includeArchived &&
		(project.title.toLowerCase().includes(normalized) ||
			project.description.toLowerCase().includes(normalized)),
	)
}

// One observer and one derived list set for the dozens of consumers.
const useSharedProjects = createSharedComposable(() => {
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
	const rootProjects = computed(() => sortByPosition(realProjects.value.filter(project =>
		!project.is_archived && (project.parent_project_id === 0 || isOrphaned(realProjects.value, project)),
	)))
	const favoriteProjects = computed(() => [
		...(favoriteProject.value ? [favoriteProject.value] : []),
		...savedFilterProjects.value.filter(project => !project.is_archived && project.is_favorite),
		...sortByPosition(realProjects.value.filter(project => !project.is_archived && project.is_favorite)),
	])
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

export function useProjects() {
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
	} = useSharedProjects()

	const state = reactive({
		projects,
		projectsArray,
		notArchivedRootProjects: rootProjects,
		favoriteProject,
		savedFilterProjects,
		favoriteProjects,
		hasProjects,
		isLoading: query.isPending,
		getChildProjects: (id: number) =>
			sortByPosition(realProjects.value.filter(project => project.parent_project_id === id)),
		getAncestors: (project: ProjectResponse) => getAncestors(realProjects.value, project),
		getEffectiveParentProjectId: (project: ProjectResponse, parentProjectIdFromDom: number) =>
			parentProjectIdFromDom === 0 && isOrphaned(realProjects.value, project)
				? project.parent_project_id
				: parentProjectIdFromDom,
		findProjectByExactname: (title: string) => findProjectByExactTitle(realProjects.value, title),
		searchProject: (value: string, includeArchived = false) =>
			search(realProjects.value, value, includeArchived),
		searchSavedFilter: (value: string, includeArchived = false) =>
			search(savedFilterProjects.value, value, includeArchived),
		searchProjectAndFilter: (value: string, includeArchived = false) =>
			search(projectsArray.value, value, includeArchived),
		loadProject: ensureProject,
	})

	return state
}
