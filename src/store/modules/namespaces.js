import Vue from 'vue'

import NamespaceService from '../../services/namespace'

export default {
	namespaced: true,
	state: () => ({
		namespaces: [],
	}),
	mutations: {
		namespaces(state, namespaces) {
			state.namespaces = namespaces
		},
		setNamespaceById(state, namespace) {
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === namespace.id) {
					namespace.lists = state.namespaces[n].lists
					Vue.set(state.namespaces, n, namespace)
					return
				}
			}
		},
		setListInNamespaceById(state, list) {
			for (const n in state.namespaces) {
				// We don't have the namespace id on the list which means we need to loop over all lists until we find it.
				// FIXME: Not ideal at all - we should fix that at the api level.
				for (const l in state.namespaces[n].lists) {
					if (state.namespaces[n].lists[l].id === list.id) {
						const namespace = state.namespaces[n]
						namespace.lists[l] = list
						Vue.set(state.namespaces, n, namespace)
						return
					}
				}
			}
		},
		addNamespace(state, namespace) {
			state.namespaces.push(namespace)
		},
		addListToNamespace(state, list) {
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === list.namespaceId) {
					state.namespaces[n].lists.push(list)
					return
				}
			}
		},
	},
	getters: {
		getListAndNamespaceById: state => listId => {
			for (const n in state.namespaces) {
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
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === namespaceId) {
					return state.namespaces[n]
				}
			}
			return null
		},
	},
	actions: {
		loadNamespaces(ctx) {
			const namespaceService = new NamespaceService()
			// We always load all namespaces and filter them on the frontend
			return namespaceService.getAll({}, {is_archived: true})
				.then(r => {
					ctx.commit('namespaces', r)

					// Put all lists in the list state
					const lists = []
					r.forEach(n => {
						n.lists.forEach(l => {
							lists.push(l)
						})
					})

					ctx.commit('lists/addLists', lists, {root: true})

					return Promise.resolve()
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
		loadNamespacesIfFavoritesDontExist(ctx) {
			// The first namespace should be the one holding all favorites
			if(ctx.state.namespaces[0].id !== -2) {
				return ctx.dispatch('loadNamespaces')
			}
		},
	},
}