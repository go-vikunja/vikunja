import {computed, readonly, ref} from 'vue'
import {defineStore, acceptHMRUpdate} from 'pinia'

import NamespaceService from '../services/namespace'
import {setModuleLoading} from '@/stores/helper'
import {createNewIndexer} from '@/indexes'
import type {INamespace} from '@/modelTypes/INamespace'
import type {IList} from '@/modelTypes/IList'
import {useListStore} from '@/stores/lists'

const {add, remove, search, update} = createNewIndexer('namespaces', ['title', 'description'])

export const useNamespaceStore = defineStore('namespace', () => {
	const listStore = useListStore()

	const isLoading = ref(false)
	// FIXME: should be object with id as key
	const namespaces = ref<INamespace[]>([])

	const getListAndNamespaceById = computed(() => (listId: IList['id'], ignorePseudoNamespaces = false) => {
		for (const n in namespaces.value) {

			if (ignorePseudoNamespaces && namespaces.value[n].id < 0) {
				continue
			}

			for (const l in namespaces.value[n].lists) {
				if (namespaces.value[n].lists[l].id === listId) {
					return {
						list: namespaces.value[n].lists[l],
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

			// Check for each list in that namespace if it has a subscription and set it if not
			n.lists.forEach(l => {
				if (l.subscription === null || l.subscription.entity !== 'list') {
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

		if (!namespace.lists || namespace.lists.length === 0) {
			namespace.lists = namespaces.value[namespaceIndex].lists
		}
		
		// Check for each list in that namespace if it has a subscription and set it if not
		namespace.lists.forEach(l => {
			if (l.subscription === null || l.subscription.entity !== 'list') {
				l.subscription = namespace.subscription
			}
		})

		namespaces.value[namespaceIndex] = namespace
		update(namespace)
	}

	function setListInNamespaceById(list: IList) {
		for (const n in namespaces.value) {
			// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
			// FIXME: Not ideal at all - we should fix that at the api level.
			if (namespaces.value[n].id === list.namespaceId) {
				for (const l in namespaces.value[n].lists) {
					if (namespaces.value[n].lists[l].id === list.id) {
						const namespace = namespaces.value[n]
						namespace.lists[l] = list
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

	function addListToNamespace(list: IList) {
		for (const n in namespaces.value) {
			if (namespaces.value[n].id === list.namespaceId) {
				namespaces.value[n].lists.push(list)
				return
			}
		}
	}

	function removeListFromNamespaceById(list: IList) {
		for (const n in namespaces.value) {
			// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
			// FIXME: Not ideal at all - we should fix that at the api level.
			if (namespaces.value[n].id === list.namespaceId) {
				for (const l in namespaces.value[n].lists) {
					if (namespaces.value[n].lists[l].id === list.id) {
						namespaces.value[n].lists.splice(l, 1)
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

			// Put all lists in the list state
			const lists = namespaces.flatMap(({lists}) => lists)

			listStore.setLists(lists)

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
		if (namespaces.value[0].id === -2 && namespaces.value[0].lists.length === 0) {
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

		getListAndNamespaceById,
		getNamespaceById,
		searchNamespace,

		setNamespaces,
		setNamespaceById,
		setListInNamespaceById,
		addNamespace,
		removeNamespaceById,
		addListToNamespace,
		removeListFromNamespaceById,
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