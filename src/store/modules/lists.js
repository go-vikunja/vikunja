import Vue from 'vue'
import ListService from '@/services/list'

const FavoriteListsNamespace = -2

export default {
	namespaced: true,
	// The state is an object which has the list ids as keys.
	state: () => ({}),
	mutations: {
		addList(state, list) {
			Vue.set(state, list.id, list)
		},
		addLists(state, lists) {
			lists.forEach(l => {
				Vue.set(state, l.id, l)
			})
		},
	},
	getters: {
		getListById: state => id => {
			if (typeof state[id] !== 'undefined') {
				return state[id]
			}
			return null
		},
	},
	actions: {
		toggleListFavorite(ctx, list) {
			list.isFavorite = !list.isFavorite
			const listService = new ListService()

			return listService.update(list)
				.then(r => {
					if (r.isFavorite) {
						ctx.commit('addList', r)
						r.namespaceId = FavoriteListsNamespace
						ctx.commit('namespaces/addListToNamespace', r, {root: true})
					} else {
						ctx.commit('namespaces/setListInNamespaceById', r, {root: true})
						r.namespaceId = FavoriteListsNamespace
						ctx.commit('namespaces/removeListFromNamespaceById', r, {root: true})
					}
					ctx.dispatch('namespaces/loadNamespacesIfFavoritesDontExist', null, {root: true})
					ctx.dispatch('namespaces/removeFavoritesNamespaceIfEmpty', null, {root: true})
					return Promise.resolve(r)
				})
				.catch(e => {
					// Reset the list state to the initial one to avoid confusion for the user
					list.isFavorite = !list.isFavorite
					ctx.commit('addList', list)
					return Promise.reject(e)
				})
		},
	},
}