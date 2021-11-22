import NamespaceService from '../../services/namespace'
import {setLoading} from '@/store/helper'
import {createNewIndexer} from '@/indexes'

const {add, remove, search, update} = createNewIndexer('namespaces', ['title', 'description'])

export default {
	namespaced: true,
	state: () => ({
		namespaces: [],
	}),
	mutations: {
		namespaces(state, namespaces) {
			state.namespaces = namespaces
			namespaces.forEach(n => {
				add(n)
			})
		},
		setNamespaceById(state, namespace) {
			const namespaceIndex = state.namespaces.findIndex(n => n.id === namespace.id)

			if (namespaceIndex === -1) {
				return
			}

			// FIXME: direct manipulation of the prop
			// might not be a problem since this is happening in the mutation 
			if (!namespace.lists || namespace.lists.length === 0) {
				namespace.lists = state.namespaces[namespaceIndex].lists
			}

			state.namespaces[namespaceIndex] = namespace
			update(namespace)
		},
		setListInNamespaceById(state, list) {
			for (const n in state.namespaces) {
				// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
				// FIXME: Not ideal at all - we should fix that at the api level.
				if (state.namespaces[n].id === list.namespaceId) {
					for (const l in state.namespaces[n].lists) {
						if (state.namespaces[n].lists[l].id === list.id) {
							const namespace = state.namespaces[n]
							namespace.lists[l] = list
							state.namespaces[n] = namespace
							return
						}
					}
				}
			}
		},
		addNamespace(state, namespace) {
			state.namespaces.push(namespace)
			add(namespace)
		},
		removeNamespaceById(state, namespaceId) {
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === namespaceId) {
					remove(state.namespaces[n])
					state.namespaces.splice(n, 1)
					return
				}
			}
		},
		addListToNamespace(state, list) {
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === list.namespaceId) {
					state.namespaces[n].lists.push(list)
					return
				}
			}
		},
		removeListFromNamespaceById(state, list) {
			for (const n in state.namespaces) {
				// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
				// FIXME: Not ideal at all - we should fix that at the api level.
				if (state.namespaces[n].id === list.namespaceId) {
					for (const l in state.namespaces[n].lists) {
						if (state.namespaces[n].lists[l].id === list.id) {
							state.namespaces[n].lists.splice(l, 1)
							return
						}
					}
				}
			}
		},
	},
	getters: {
		getListAndNamespaceById: state => (listId, ignorePseudoNamespaces = false) => {
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
		getNamespaceById: state => namespaceId => {
			return state.namespaces.find(({id}) => id == namespaceId) || null
		},
		searchNamespace: (state, getters) => query => {
			return search(query)
					?.filter(value => value > 0)
					.map(getters.getNamespaceById)
					.filter(n => n !== null)
				|| []
		},
	},
	actions: {
		async loadNamespaces(ctx) {
			const cancel = setLoading(ctx, 'namespaces')

			const namespaceService = new NamespaceService()
			try {
				// We always load all namespaces and filter them on the frontend
				const namespaces = await namespaceService.getAll({}, {is_archived: true})
				ctx.commit('namespaces', namespaces)

				// Put all lists in the list state
				const lists = namespaces.flatMap(({lists}) => lists)

				ctx.commit('lists/setLists', lists, {root: true})

				return namespaces
			} finally {
				cancel()
			}
		},

		loadNamespacesIfFavoritesDontExist(ctx) {
			// The first namespace should be the one holding all favorites
			if (ctx.state.namespaces[0].id !== -2) {
				return ctx.dispatch('loadNamespaces')
			}
		},

		removeFavoritesNamespaceIfEmpty(ctx) {
			if (ctx.state.namespaces[0].id === -2 && ctx.state.namespaces[0].lists.length === 0) {
				ctx.state.namespaces.splice(0, 1)
			}
		},

		async deleteNamespace(ctx, namespace) {
			const cancel = setLoading(ctx, 'namespaces')
			const namespaceService = new NamespaceService()

			try {
				const response = await namespaceService.delete(namespace)
				ctx.commit('removeNamespaceById', namespace.id)
				return response
			} finally {
				cancel()
			}
		},

		async createNamespace(ctx, namespace) {
			const cancel = setLoading(ctx, 'namespaces')
			const namespaceService = new NamespaceService()

			try {
				const createdNamespace = await namespaceService.create(namespace)
				ctx.commit('addNamespace', createdNamespace)
				return createdNamespace
			} finally {
				cancel()
			}
		},
	},
}