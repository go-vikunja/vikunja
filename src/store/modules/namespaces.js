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
	actions: {
		loadNamespaces(ctx) {
			const namespaceService = new NamespaceService()
			// We always load all namespaces and filter them on the frontend
			return namespaceService.getAll({}, {is_archived: true})
				.then(r => {
					ctx.commit('namespaces', r)
					return Promise.resolve()
				})
				.catch(e => {
					return Promise.reject(e)
				})
		},
	},
}