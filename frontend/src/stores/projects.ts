import {watch, reactive, shallowReactive, toValue, readonly, ref, computed, type MaybeRefOrGetter} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import ProjectService from '@/services/project'
import ProjectDuplicateService from '@/services/projectDuplicateService'
import ProjectDuplicateModel from '@/models/projectDuplicateModel'
import {setModuleLoading} from '@/stores/helper'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {createNewIndexer} from '@/indexes'

import type {IProject} from '@/modelTypes/IProject'

import ProjectModel from '@/models/project'
import {success} from '@/message'
import {useBaseStore} from '@/stores/base'
import {getSavedFilterIdFromProjectId} from '@/services/savedFilter'
import type {IProjectView} from '@/modelTypes/IProjectView'
import {PERMISSIONS} from '@/constants/permissions.ts'

const {add, remove, search, update} = createNewIndexer('projects', ['title', 'description'])

export const useProjectStore = defineStore('project', () => {
	const baseStore = useBaseStore()
	const router = useRouter()

	const isLoading = ref(false)

	// The projects are stored as an object which has the project ids as keys.
	const projects = ref<{ [id: IProject['id']]: IProject }>({})
	const projectsArray = computed(() => Object.values(projects.value)
		.sort((a, b) => a.position - b.position))

	const notArchivedRootProjects = computed(() => projectsArray.value
		.filter(p => p.parentProjectId === 0 && !p.isArchived && p.id > 0))
	const favoriteProjects = computed(() => projectsArray.value
		.filter(p => !p.isArchived && p.isFavorite))
	const savedFilterProjects = computed(() => projectsArray.value
		.filter(p => !p.isArchived && p.id < -1))
	const hasProjects = computed(() => projectsArray.value.length > 0)

	const getChildProjects = computed(() => {
		return (id: IProject['id']) => projectsArray.value.filter(p => p.parentProjectId === id)
	})

	const getAncestors = computed(() => {
		return (project: IProject): IProject[] => {
			if (typeof project === 'undefined') {
				return []
			}

			if (!project?.parentProjectId) {
				return [project]
			}

			const parentProject = projects.value[project.parentProjectId]
			return [
				...(parentProject ? getAncestors.value(parentProject) : []),
				project,
			]
		}
	})

	const findProjectByExactname = computed(() => {
		return (name: string) => {
			const project = projectsArray.value.find(l => {
				return l.title.toLowerCase() === name.toLowerCase()
			})
			return typeof project === 'undefined' ? null : project
		}
	})
	
	const findProjectByIdentifier = computed(() => {
		return (identifier: string) => {
			const project = projectsArray.value.find(p => {
				return p.identifier.toLowerCase() === identifier.toLowerCase()
			})
			return typeof project === 'undefined' ? null : project
		}
	})

	const searchProjectAndFilter = computed(() => {
		return (query: string, includeArchived = false) => {
			return search(query)
				?.map(id => projects.value[id])
				.filter(project => project?.isArchived === includeArchived)
				|| []
		}
	})

	const searchProject = computed(() => {
		return (query: string, includeArchived = false) => {
			return search(query)
				?.filter(value => value > 0)
				.map(id => projects.value[id])
				.filter(project => project?.isArchived === includeArchived)
			|| []
		}
	})
	
	const searchSavedFilter = computed(() => {
		return (query: string, includeArchived = false) => {
			return search(query)
				?.filter(value => getSavedFilterIdFromProjectId(value) > 0)
				.map(id => projects.value[id])
				.filter(project => project?.isArchived === includeArchived)
				|| []
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setProject(project: IProject) {
		projects.value[project.id] = project
		update(project)

		// FIXME: This should be a watcher, but using a watcher instead will sometimes crash browser processes.
		// Reverted from 31b7c1f217532bf388ba95a03f469508bee46f6a
		if (baseStore.currentProject?.id === project.id) {
			baseStore.setCurrentProject(project)
		}
	}

	function setProjects(newProjects: IProject[]) {
		newProjects.forEach(p => setProject(p))
	}

	function removeProjectById(project: IProject) {
		
		// Remove child projects from state as well
		projectsArray.value
			.filter(p => p.parentProjectId === project.id)
			.forEach(p => removeProjectById(p))
		
		remove(project)
		delete projects.value[project.id]
	}

	function toggleProjectFavorite(project: IProject) {
		// The favorites pseudo project is always favorite
		// Archived projects cannot be marked favorite
		if (project.id === -1 || project.isArchived) {
			return
		}
		return updateProject({
			...project,
			isFavorite: !project.isFavorite,
		})
	}

	async function createProject(project: IProject) {
		const cancel = setModuleLoading(setIsLoading)
		const projectService = new ProjectService()

		try {
			const createdProject = await projectService.create(project)
			setProject(createdProject)
			router.push({
				name: 'project.index',
				params: { projectId: createdProject.id },
			})
			return createdProject
		} finally {
			cancel()
		}
	}

	async function updateProject(project: IProject) {
		const cancel = setModuleLoading(setIsLoading)
		const projectService = new ProjectService()
		
		try {
			const updatedProject = await projectService.update(project)
			setProject(project)

			// the returned project from projectService.update is the same!
			// in order to not create a manipulation in pinia store we have to create a new copy
			return updatedProject
		} catch (e) {
			// Reset the project state to the initial one to avoid confusion for the user
			setProject({
				...project,
				isFavorite: !project.isFavorite,
			})
			throw e
		} finally {
			cancel()
		}
	}

	async function deleteProject(project: IProject) {
		const cancel = setModuleLoading(setIsLoading)
		const projectService = new ProjectService()

		try {
			const response = await projectService.delete(project)
			removeProjectById(project)
			removeProjectFromHistory({id: project.id})
			return response
		} finally {
			cancel()
		}
	}

	async function loadAllProjects() {
		const cancel = setModuleLoading(setIsLoading)

		const projectService = new ProjectService()
		const loadedProjects: IProject[] = []
		let page = 1
		try {
			do {
				const newProjects = await projectService.getAll({}, {is_archived: true, expand: 'permissions'}, page) as IProject[]
				loadedProjects.push(...newProjects)
				page++
			} while (page <= projectService.totalPages)
			
		} finally {
			cancel()
		}
		
		projects.value = {}
		setProjects(loadedProjects)
		loadedProjects.forEach(p => add(p))

		return loadedProjects
	}
	
	function setProjectView(view: IProjectView) {
		const views = [...projects.value[view.projectId].views]
		const viewPos = views.findIndex(v => v.id === view.id)

		if (viewPos !== -1) {
			views[viewPos] = view
		} else {
			views.push(view)
		}
		views.sort((a, b) => a.position - b.position)
		
		setProject({
			...projects.value[view.projectId],
			views,
		})
	}
	
	function removeProjectView(projectId: IProject['id'], viewId: IProjectView['id']) {
		const project = projects.value[projectId]
		const updatedViews = project.views.filter(v => v.id !== viewId)
	
		setProject({
			...project,
			views: updatedViews,
		})
	}
	
	return {
		isLoading: readonly(isLoading),
		projects: readonly(projects),
		projectsArray: readonly(projectsArray),
		notArchivedRootProjects: readonly(notArchivedRootProjects),
		favoriteProjects: readonly(favoriteProjects),
		hasProjects: readonly(hasProjects),
		savedFilterProjects: readonly(savedFilterProjects),

		getChildProjects,
		findProjectByExactname,
		findProjectByIdentifier,
		searchProject,
		searchSavedFilter,
		searchProjectAndFilter,

		setProject,
		setProjects,
		removeProjectById,
		toggleProjectFavorite,
		loadAllProjects,
		createProject,
		updateProject,
		deleteProject,
		getAncestors,
		setProjectView,
		removeProjectView,
	}
})

export function useProject(projectId: MaybeRefOrGetter<IProject['id']>) {
	const projectService = shallowReactive(new ProjectService())
	const projectDuplicateService = shallowReactive(new ProjectDuplicateService())
	
	const isLoading = computed(() => projectService.loading || projectDuplicateService.loading)
	const project: IProject = reactive(new ProjectModel())
	
	const {t} = useI18n({useScope: 'global'})
	const router = useRouter()
	const projectStore = useProjectStore()

	watch(
		() => toValue(projectId),
		async (projectId) => {
			const loadedProject = await projectService.get(new ProjectModel({id: projectId}))
			Object.assign(project, loadedProject)
		},
		{immediate: true},
	)

	async function save() {
		const updatedProject = await projectStore.updateProject(project)
		Object.assign(project, updatedProject)
		success({message: t('project.edit.success')})
	}
	
	async function duplicateProject(parentProjectId: IProject['id']) {
		const projectDuplicate = new ProjectDuplicateModel({
			projectId: Number(toValue(projectId)),
			parentProjectId,
		})

		const duplicate = await projectDuplicateService.create(projectDuplicate)
		if (duplicate.duplicatedProject) {
			duplicate.duplicatedProject.maxPermission = PERMISSIONS.ADMIN
		}

		projectStore.setProject(duplicate.duplicatedProject)
		success({message: t('project.duplicate.success')})
		router.push({name: 'project.index', params: {projectId: duplicate.duplicatedProject.id}})
	}

	return {
		isLoading: readonly(isLoading),
		project,
		save,
		duplicateProject,
	}
}

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useProjectStore, import.meta.hot))
}
