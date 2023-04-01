import {watch, reactive, shallowReactive, unref, toRefs, readonly, ref, computed} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {useI18n} from 'vue-i18n'

import ProjectService from '@/services/project'
import {setModuleLoading} from '@/stores/helper'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {createNewIndexer} from '@/indexes'

import type {IProject} from '@/modelTypes/IProject'

import type {MaybeRef} from '@vueuse/core'

import ProjectModel from '@/models/project'
import {success} from '@/message'
import {useBaseStore} from '@/stores/base'

const {add, remove, search, update} = createNewIndexer('projects', ['title', 'description'])

export interface ProjectState {
	[id: IProject['id']]: IProject
}

export const useProjectStore = defineStore('project', () => {
	const baseStore = useBaseStore()

	const isLoading = ref(false)

	// The projects are stored as an object which has the project ids as keys.
	const projects = ref<ProjectState>({})
	const projectsArray = computed(() => Object.values(projects.value))
	const notArchivedRootProjects = computed(() => projectsArray.value
		.filter(p => p.parentProjectId === 0 && !p.isArchived))
	const hasProjects = computed(() => projects.value ? true : false)

	const getProjectById = computed(() => {
		return (id: IProject['id']) => typeof projects.value[id] !== 'undefined' ? projects.value[id] : null
	})
	const getChildProjects = computed(() => {
		return (id: IProject['id']) => projectsArray.value.filter(p => p.parentProjectId === id) || []
	})

	const findProjectByExactname = computed(() => {
		return (name: string) => {
			const project = Object.values(projects.value).find(l => {
				return l.title.toLowerCase() === name.toLowerCase()
			})
			return typeof project === 'undefined' ? null : project
		}
	})

	const searchProject = computed(() => {
		return (query: string, includeArchived = false) => {
			return search(query)
					?.filter(value => value > 0)
					.map(id => projects.value[id])
					.filter(project => project.isArchived === includeArchived)
				|| []
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setProject(project: IProject, updateChildren: boolean = true) {
		projects.value[project.id] = project
		update(project)

		if (baseStore.currentProject?.id === project.id) {
			baseStore.setCurrentProject(project)
		}

		if (updateChildren) {
			// When projects are loaded from the api, they will include child projects
			// in the `childProjects` property. We flatten them out into the project store here.
			project.childProjects?.forEach(p => setProject(new ProjectModel(p)))
			delete project.childProjects
		}

		// if the project is a child project, we need to also set it in the parent
		if (project.parentProjectId) {
			const parent = projects.value[project.parentProjectId]
			let foundProject = false
			parent.childProjectIds = parent.childProjectIds?.forEach(p => {
				if (p.id === project.id) {
					foundProject = true
				}
			})
			// If we hit this, the project now has a new parent project which it did not have before
			if (!foundProject) {
				if (!parent.childProjectIds) {
					parent.childProjectIds = []
				}
				parent.childProjectIds.push(project.id)
			}
			setProject(parent, false)
		}
	}

	function setProjects(newProjects: IProject[]) {
		newProjects.forEach(p => setProject(p))
	}

	function removeProjectById(project: IProject) {
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
			return createdProject
		} finally {
			cancel()
		}
	}

	async function updateProject(project: IProject) {
		const cancel = setModuleLoading(setIsLoading)
		const projectService = new ProjectService()

		try {
			await projectService.update(project)
			setProject(project)

			// the returned project from projectService.update is the same!
			// in order to not create a manipulation in pinia store we have to create a new copy
			return {
				...project,
			}
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

	async function loadProjects() {
		const cancel = setModuleLoading(setIsLoading)

		const projectService = new ProjectService()
		try {
			const loadedProjects = await projectService.getAll({}, {is_archived: true}) as IProject[]
			projects.value = {}
			setProjects(loadedProjects)

			return loadedProjects
		} finally {
			cancel()
		}
	}

	function getParentProjects(project: IProject): IProject[] {
		if (!project?.parentProjectId) {
			return [project]
		}

		const parentProject = projects.value[project.parentProjectId]
		return [
			...getParentProjects(parentProject),
			project,
		]
	}

	return {
		isLoading: readonly(isLoading),
		projects: readonly(projects),
		projectsArray: readonly(projectsArray),
		notArchivedRootProjects: readonly(notArchivedRootProjects),
		hasProjects: readonly(hasProjects),

		getProjectById,
		getChildProjects,
		findProjectByExactname,
		searchProject,

		setProject,
		setProjects,
		removeProjectById,
		toggleProjectFavorite,
		loadProjects,
		createProject,
		updateProject,
		deleteProject,
		getParentProjects,
	}
})

export function useProject(projectId: MaybeRef<IProject['id']>) {
	const projectService = shallowReactive(new ProjectService())
	const {loading: isLoading} = toRefs(projectService)
	const project: IProject = reactive(new ProjectModel())
	const {t} = useI18n({useScope: 'global'})

	watch(
		() => unref(projectId),
		async (projectId) => {
			const loadedProject = await projectService.get(new ProjectModel({id: projectId}))
			Object.assign(project, loadedProject)
		},
		{immediate: true},
	)

	const projectStore = useProjectStore()

	const parentProject = ref<IProject | null>(null)
	watch(
		() => project.parentProjectId,
		projectId => {
			if (project.parentProjectId) {
				parentProject.value = projectStore.getProjectById(project.parentProjectId)
			}
		},
		{immediate: true},
	)

	async function save() {
		project.parentProjectId = parentProject.value.id
		await projectStore.updateProject(project)
		success({message: t('project.edit.success')})
	}

	return {
		isLoading: readonly(isLoading),
		project,
		parentProject,
		save,
	}
}

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useProjectStore, import.meta.hot))
}