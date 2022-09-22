import {defineStore, acceptHMRUpdate} from 'pinia'

import NamespaceService from '../services/namespace'
import {setLoadingPinia} from '@/store/helper'
import {createNewIndexer} from '@/indexes'
import type {NamespaceState} from '@/store/types'
import type {INamespace} from '@/modelTypes/INamespace'
import type {IList} from '@/modelTypes/IList'
import {useListStore} from '@/stores/lists'

const {add, remove, search, update} = createNewIndexer('namespaces', ['title', 'description'])

export const useNamespaceStore = defineStore('namespace', {
	state: (): NamespaceState => ({
		isLoading: false,
		// FIXME: should be object with id as key
		namespaces: [],
	}),
	getters: {
		getListAndNamespaceById: (state) => (listId: IList['id'], ignorePseudoNamespaces = false) => {
			for (const n in state.namespaces) {

				if (ignorePseudoNamespaces && state.namespaces[n].id < 0) {
					continue
				}

				for (const l in state.namespaces[n].lists) {
					if (state.namespaces[n].lists[l].id === listId) {
						return {
							list: state.namespaces[n].lists[l],
							namespace: state.namespaces[n],
						}
					}
				}
			}
			return null
		},

		getNamespaceById: state => (namespaceId: INamespace['id']) => {
			return state.namespaces.find(({id}) => id == namespaceId) || null
		},
	
		searchNamespace() {
			return (query: string) => (
				search(query)
					?.filter(value => value > 0)
					.map(this.getNamespaceById)
					.filter(n => n !== null)
				|| []
			)
		},
	},

	actions: {
		setIsLoading(isLoading: boolean) {
			this.isLoading = isLoading
		},

		setNamespaces(namespaces: INamespace[]) {
			this.namespaces = namespaces
			namespaces.forEach(n => {
				add(n)
			})
		},

		setNamespaceById(namespace: INamespace) {
			const namespaceIndex = this.namespaces.findIndex(n => n.id === namespace.id)

			if (namespaceIndex === -1) {
				return
			}

			if (!namespace.lists || namespace.lists.length === 0) {
				namespace.lists = this.namespaces[namespaceIndex].lists
			}

			this.namespaces[namespaceIndex] = namespace
			update(namespace)
		},

		setListInNamespaceById(list: IList) {
			for (const n in this.namespaces) {
				// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
				// FIXME: Not ideal at all - we should fix that at the api level.
				if (this.namespaces[n].id === list.namespaceId) {
					for (const l in this.namespaces[n].lists) {
						if (this.namespaces[n].lists[l].id === list.id) {
							const namespace = this.namespaces[n]
							namespace.lists[l] = list
							this.namespaces[n] = namespace
							return
						}
					}
				}
			}
		},

		addNamespace(namespace: INamespace) {
			this.namespaces.push(namespace)
			add(namespace)
		},

		removeNamespaceById(namespaceId: INamespace['id']) {
			for (const n in this.namespaces) {
				if (this.namespaces[n].id === namespaceId) {
					remove(this.namespaces[n])
					this.namespaces.splice(n, 1)
					return
				}
			}
		},

		addListToNamespace(list: IList) {
			for (const n in this.namespaces) {
				if (this.namespaces[n].id === list.namespaceId) {
					this.namespaces[n].lists.push(list)
					return
				}
			}
		},

		removeListFromNamespaceById(list: IList) {
			for (const n in this.namespaces) {
				// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
				// FIXME: Not ideal at all - we should fix that at the api level.
				if (this.namespaces[n].id === list.namespaceId) {
					for (const l in this.namespaces[n].lists) {
						if (this.namespaces[n].lists[l].id === list.id) {
							this.namespaces[n].lists.splice(l, 1)
							return
						}
					}
				}
			}
		},

		async loadNamespaces() {
			const cancel = setLoadingPinia(this)

			const namespaceService = new NamespaceService()
			try {
				// We always load all namespaces and filter them on the frontend
				const namespaces = await namespaceService.getAll({}, {is_archived: true}) as INamespace[]
				this.setNamespaces(namespaces)

				// Put all lists in the list state
				const lists = namespaces.flatMap(({lists}) => lists)

				const listStore = useListStore()
				listStore.setLists(lists)

				return namespaces
			} finally {
				cancel()
			}
		},

		loadNamespacesIfFavoritesDontExist() {
			// The first or second namespace should be the one holding all favorites
			if (this.namespaces[0].id === -2 || this.namespaces[1]?.id === -2) {
				return
			}
			return this.loadNamespaces()
		},

		removeFavoritesNamespaceIfEmpty() {
			if (this.namespaces[0].id === -2 && this.namespaces[0].lists.length === 0) {
				this.namespaces.splice(0, 1)
			}
		},

		async deleteNamespace(namespace: INamespace) {
			const cancel = setLoadingPinia(this)
			const namespaceService = new NamespaceService()

			try {
				const response = await namespaceService.delete(namespace)
				this.removeNamespaceById(namespace.id)
				return response
			} finally {
				cancel()
			}
		},

		async createNamespace(namespace: INamespace) {
			const cancel = setLoadingPinia(this)
			const namespaceService = new NamespaceService()

			try {
				const createdNamespace = await namespaceService.create(namespace)
				this.addNamespace(createdNamespace)
				return createdNamespace
			} finally {
				cancel()
			}
		},
	},
})

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useNamespaceStore, import.meta.hot))
}