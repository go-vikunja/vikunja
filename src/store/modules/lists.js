import ListService from '@/services/list'
import {setLoading} from '@/store/helper'
import {removeListFromHistory} from '@/modules/listHistory.ts'

const FavoriteListsNamespace = -2

export default {
	namespaced: true,
	// The state is an object which has the list ids as keys.
	state: () => ({}),
	mutations: {
		setList(state, list) {
			state[list.id] = list
		},
		setLists(state, lists) {
			lists.forEach(l => {
				state[l.id] = l
			})
		},
		removeListById(state, list) {
			delete state[list.id]
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
			return ctx.dispatch('updateList', {
				...list,
				isFavorite: !list.isFavorite,
			})
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
				.finally(() => cancel())
		},
		updateList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			return listService.update(list)
				.then(() => {
					ctx.commit('setList', list)
					ctx.commit('namespaces/setListInNamespaceById', list, {root: true})

					// the returned list from listService.update is the same!
					// in order to not validate vuex mutations we have to create a new copy
					const newList = {
						...list,
						namespaceId: FavoriteListsNamespace,
					}
					if (list.isFavorite) {
						ctx.commit('namespaces/addListToNamespace', newList, {root: true})
					} else {
						ctx.commit('namespaces/removeListFromNamespaceById', newList, {root: true})
					}
					ctx.dispatch('namespaces/loadNamespacesIfFavoritesDontExist', null, {root: true})
					ctx.dispatch('namespaces/removeFavoritesNamespaceIfEmpty', null, {root: true})
					return Promise.resolve(newList)
				})
				.catch(e => {
					// Reset the list state to the initial one to avoid confusion for the user
					ctx.commit('setList', {
						...list,
						isFavorite: !list.isFavorite,
					})
					return Promise.reject(e)
				})
				.finally(() => cancel())
		},
		deleteList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			return listService.delete(list)
				.then(r => {
					ctx.commit('removeListById', list)
					ctx.commit('namespaces/removeListFromNamespaceById', list, {root: true})
					removeListFromHistory({id: list.id})
					return r
				})
				.finally(() => cancel())
		},
	},
}