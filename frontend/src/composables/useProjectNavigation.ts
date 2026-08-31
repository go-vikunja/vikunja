import {computed, reactive} from 'vue'
import {useQuery} from '@tanstack/vue-query'
import {useRouter} from 'vue-router'

import {
	createProject,
	deleteProject,
	findProjectByExactTitle,
	findProjectByIdentifier,
	getChildProjects,
	getEffectiveParentProjectId,
	getFavoriteNavigationItems,
	getProjectAncestors,
	getRootProjects,
	getSavedFilterIdFromProjectId,
	isSavedFilterProject,
	normalizeProject,
	patchProjectFavorite,
	projectKeys,
	projectQuery,
	projectsQuery,
	refreshProjects,
	searchProjects,
	updateProject,
	updateProjectInCache,
	updateProjectNavigationItemInCache,
} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import type {Project, ProjectView, ProjectWritable} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import SavedFilterModel from '@/models/savedFilter'
import SavedFilterService from '@/services/savedFilter'

export function useProjectNavigation() {
	const router = useRouter()
	const query = useQuery(projectsQuery())
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

	async function toggleProjectFavorite(projectId: number) {
		if (projectId === -1) {
			return
		}

		const project = projectId > 0
			? realProjects.value.find(project => project.id === projectId)
			: savedFilterProjects.value.find(project => project.id === projectId)
		if (!project || project.is_archived) {
			return
		}

		if (!isSavedFilterProject(project)) {
			await patchProjectFavorite(project.id, !project.is_favorite)
			return
		}

		const previous = project.is_favorite
		updateProjectNavigationItemInCache(project.id, current => ({
			...current,
			is_favorite: !previous,
		}))
		try {
			const filterId = getSavedFilterIdFromProjectId(project.id)
			if (!filterId) {
				return
			}
			const service = new SavedFilterService()
			const filter = await service.get(new SavedFilterModel({id: filterId}))
			filter.isFavorite = !previous
			await service.update(filter)
		} catch (error) {
			updateProjectNavigationItemInCache(project.id, current => ({
				...current,
				is_favorite: previous,
			}))
			throw error
		}
	}

	async function create(project: ProjectWritable) {
		const created = await createProject(project)
		await router.push({name: 'project.index', params: {projectId: created.id}})
		return created
	}

	async function remove(project: Pick<Project, 'id'>) {
		if (typeof project.id !== 'number') {
			return
		}
		await deleteProject(project.id)
	}

	function setProject(project: Project) {
		const normalized = normalizeProject(project)
		const existing = projects.value[normalized.id]
		if (existing) {
			updateProjectNavigationItemInCache(normalized.id, () => normalized)
			updateProjectInCache(normalized.id, () => normalized)
			return
		}
		queryClient.setQueryData(projectQuery(normalized.id).queryKey, normalized)
		queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	}

	function setProjectView(view: ProjectView) {
		if (typeof view.project_id !== 'number') {
			return
		}
		updateProjectInCache(view.project_id, project => ({
			...project,
			views: [...project.views.filter(current => current.id !== view.id), view]
				.sort((a, b) => (a.position ?? 0) - (b.position ?? 0)),
		}))
	}

	function removeProjectView(projectId: number, viewId: number) {
		updateProjectInCache(projectId, project => ({
			...project,
			views: project.views.filter(view => view.id !== viewId),
		}))
	}

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
		toggleProjectFavorite: (project: ProjectResponse) => toggleProjectFavorite(project.id),
		loadAllProjects: refreshProjects,
		loadProject: (projectId: number) => queryClient.ensureQueryData(projectQuery(projectId)),
		createProject: create,
		updateProject,
		deleteProject: remove,
		setProject,
		setProjectView,
		removeProjectView,
	})

	return state
}
