import {computed, readonly, ref} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'

import NamespaceService from '../services/namespace'
import {setModuleLoading} from '@/stores/helper'
import {createNewIndexer} from '@/indexes'
import type {INamespace} from '@/modelTypes/INamespace'
import type {IProject} from '@/modelTypes/IProject'
import {useProjectStore} from '@/stores/projects'

const {add, remove, search, update} = createNewIndexer('namespaces', ['title', 'description'])

export const useNamespaceStore = defineStore('namespace', () => {
	const projectStore = useProjectStore()

	const isLoading = ref(false)
	// FIXME: should be object with id as key
	const namespaces = ref<INamespace[]>([])

	const getProjectAndNamespaceById = computed(() => (projectId: IProject['id'], ignorePseudoNamespaces = false) => {
		for (const n in namespaces.value) {

			if (ignorePseudoNamespaces && namespaces.value[n].id < 0) {
				continue
			}

			for (const l in namespaces.value[n].projects) {
				if (namespaces.value[n].projects[l].id === projectId) {
					return {
						project: namespaces.value[n].projects[l],
						namespace: namespaces.value[n],
					}
				}
			}
		}
		return null
	})

	const getNamespaceById = computed(() => (namespaceId: INamespace['id']) => {
		return namespaces.value.find(({id}) => id == namespaceId) || null
	})

	const searchNamespace = computed(() => {
		return (query: string) => (
			search(query)
				?.filter(value => value > 0)
				.map(getNamespaceById.value)
				.filter(n => n !== null)
			|| []
		)
	})


	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setNamespaces(newNamespaces: INamespace[]) {
		namespaces.value = newNamespaces
		newNamespaces.forEach(n => {
			add(n)

			// Check for each project in that namespace if it has a subscription and set it if not
			n.projects.forEach(l => {
				if (l.subscription === null || l.subscription.entity !== 'project') {
					l.subscription = n.subscription
				}
			})
		})
	}

	function setNamespaceById(namespace: INamespace) {
		const namespaceIndex = namespaces.value.findIndex(n => n.id === namespace.id)

		if (namespaceIndex === -1) {
			return
		}

		if (!namespace.projects || namespace.projects.length === 0) {
			namespace.projects = namespaces.value[namespaceIndex].projects
		}
		
		// Check for each project in that namespace if it has a subscription and set it if not
		namespace.projects.forEach(l => {
			if (l.subscription === null || l.subscription.entity !== 'project') {
				l.subscription = namespace.subscription
			}
		})

		namespaces.value[namespaceIndex] = namespace
		update(namespace)
	}

	function setProjectInNamespaceById(project: IProject) {
		for (const n in namespaces.value) {
			// We don't have the namespace id on the project which means we need to loop over all projects until we find it.
			// FIXME: Not ideal at all - we should fix that at the api level.
			if (namespaces.value[n].id === project.namespaceId) {
				for (const l in namespaces.value[n].projects) {
					if (namespaces.value[n].projects[l].id === project.id) {
						const namespace = namespaces.value[n]
						namespace.projects[l] = project
						namespaces.value[n] = namespace
						return
					}
				}
			}
		}
	}

	function addNamespace(namespace: INamespace) {
		namespaces.value.push(namespace)
		add(namespace)
	}

	function removeNamespaceById(namespaceId: INamespace['id']) {
		for (const n in namespaces.value) {
			if (namespaces.value[n].id === namespaceId) {
				remove(namespaces.value[n])
				namespaces.value.splice(n, 1)
				return
			}
		}
	}

	function addProjectToNamespace(project: IProject) {
		for (const n in namespaces.value) {
			if (namespaces.value[n].id === project.namespaceId) {
				namespaces.value[n].projects.push(project)
				return
			}
		}
	}

	function removeProjectFromNamespaceById(project: IProject) {
		for (const n in namespaces.value) {
			// We don't have the namespace id on the project which means we need to loop over all projects until we find it.
			// FIXME: Not ideal at all - we should fix that at the api level.
			if (namespaces.value[n].id === project.namespaceId) {
				for (const l in namespaces.value[n].projects) {
					if (namespaces.value[n].projects[l].id === project.id) {
						namespaces.value[n].projects.splice(l, 1)
						return
					}
				}
			}
		}
	}

	async function loadNamespaces() {
		const cancel = setModuleLoading(setIsLoading)

		const namespaceService = new NamespaceService()
		try {
			// We always load all namespaces and filter them on the frontend
			const namespaces = await namespaceService.getAll({}, {is_archived: true}) as INamespace[]
			setNamespaces(namespaces)

			// Put all projects in the project state
			const projects = namespaces.flatMap(({projects}) => projects)

			projectStore.setProjects(projects)

			return namespaces
		} finally {
			cancel()
		}
	}

	function loadNamespacesIfFavoritesDontExist() {
		// The first or second namespace should be the one holding all favorites
		if (namespaces.value[0].id === -2 || namespaces.value[1]?.id === -2) {
			return
		}
		return loadNamespaces()
	}

	function removeFavoritesNamespaceIfEmpty() {
		if (namespaces.value[0].id === -2 && namespaces.value[0].projects.length === 0) {
			namespaces.value.splice(0, 1)
		}
	}

	async function deleteNamespace(namespace: INamespace) {
		const cancel = setModuleLoading(setIsLoading)
		const namespaceService = new NamespaceService()

		try {
			const response = await namespaceService.delete(namespace)
			removeNamespaceById(namespace.id)
			return response
		} finally {
			cancel()
		}
	}

	async function createNamespace(namespace: INamespace) {
		const cancel = setModuleLoading(setIsLoading)
		const namespaceService = new NamespaceService()

		try {
			const createdNamespace = await namespaceService.create(namespace)
			addNamespace(createdNamespace)
			return createdNamespace
		} finally {
			cancel()
		}
	}

	return {
		isLoading: readonly(isLoading),
		namespaces: readonly(namespaces),

		getProjectAndNamespaceById,
		getNamespaceById,
		searchNamespace,

		setNamespaces,
		setNamespaceById,
		setProjectInNamespaceById,
		addNamespace,
		removeNamespaceById,
		addProjectToNamespace,
		removeProjectFromNamespaceById,
		loadNamespaces,
		loadNamespacesIfFavoritesDontExist,
		removeFavoritesNamespaceIfEmpty,
		deleteNamespace,
		createNamespace,
	}
})

// support hot reloading
if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useNamespaceStore, import.meta.hot))
}