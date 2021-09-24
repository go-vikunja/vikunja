import Vue from 'vue'

import NamespaceService from '../../services/namespace'
import {setLoading} from '@/store/helper'

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
			const namespaceIndex = state.namespaces.findIndex(n => n.id === namespace.id)

			if (namespaceIndex === -1) {
				return
			}

			if (!namespace.lists || namespace.lists.length === 0) {
				namespace.lists = state.namespaces[namespaceIndex].lists
			}

			Vue.set(state.namespaces, namespaceIndex, namespace)
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
							Vue.set(state.namespaces, n, namespace)
							return
						}
					}
				}
			}
		},
		addNamespace(state, namespace) {
			state.namespaces.push(namespace)
		},
		removeNamespaceById(state, namespaceId) {
			for (const n in state.namespaces) {
				if (state.namespaces[n].id === namespaceId) {
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
			const cancel = setLoading(ctx, 'namespaces')

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

					ctx.commit('lists/setLists', lists, {root: true})

					return Promise.resolve(r)
				})
				.catch(e => Promise.reject(e))
				.finally(() => {
					cancel()
				})
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
				return Promise.resolve()
			}
		},
		deleteNamespace(ctx, namespace) {
			const cancel = setLoading(ctx, 'namespaces')
			const namespaceService = new NamespaceService()

			return namespaceService.delete(namespace)
				.then(r => {
					ctx.commit('removeNamespaceById', namespace.id)
					return Promise.resolve(r)
				})
				.catch(e => Promise.reject(e))
				.finally(() => cancel())
		},
		createNamespace(ctx, namespace) {
			const cancel = setLoading(ctx, 'namespaces')
			const namespaceService = new NamespaceService()

			return namespaceService.create(namespace)
				.then(r => {
					ctx.commit('addNamespace', r)
					return Promise.resolve(r)
				})
				.catch(e => Promise.reject(e))
				.finally(() => cancel())
		},
	},
}