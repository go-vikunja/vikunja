import Vue from 'vue'
import ListService from '@/services/list'
import {setLoading} from '@/store/helper'

const FavoriteListsNamespace = -2

export default {
	namespaced: true,
	// The state is an object which has the list ids as keys.
	state: () => ({}),
	mutations: {
		setList(state, list) {
			Vue.set(state, list.id, list)
		},
		setLists(state, lists) {
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
		findListByExactname: state => name => {
			const list = Object.values(state).find(l => {
				return l.title.toLowerCase() === name.toLowerCase()
			})
			return typeof list === 'undefined' ? null : list
		},
	},
	actions: {
		toggleListFavorite(ctx, list) {
			list.isFavorite = !list.isFavorite

			return ctx.dispatch('updateList', list)
		},
		createList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			return listService.create(list)
				.then(r => {
					r.namespaceId = list.namespaceId
					ctx.commit('namespaces/addListToNamespace', r, {root: true})
					ctx.commit('setList', r)
					return Promise.resolve(r)
				})
				.catch(e => Promise.reject(e))
				.finally(() => cancel())
		},
		updateList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			return listService.update(list)
				.then(r => {
					ctx.commit('setList', r)
					ctx.commit('namespaces/setListInNamespaceById', r, {root: true})
					if (r.isFavorite) {
						r.namespaceId = FavoriteListsNamespace
						ctx.commit('namespaces/addListToNamespace', r, {root: true})
					} else {
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
					ctx.commit('setList', list)
					return Promise.reject(e)
				})
				.finally(() => cancel())
		}
	},
}