import ListService from '@/services/list'
import {setLoading} from '@/store/helper'
import {removeListFromHistory} from '@/modules/listHistory.ts'
import {createNewIndexer} from '@/indexes'

const {add, remove, search, update} = createNewIndexer('lists', ['title', 'description'])

const FavoriteListsNamespace = -2

export default {
	namespaced: true,
	// The state is an object which has the list ids as keys.
	state: () => ({}),
	mutations: {
		setList(state, list) {
			state[list.id] = list
			update(list)
		},
		setLists(state, lists) {
			lists.forEach(l => {
				state[l.id] = l
				add(l)
			})
		},
		removeListById(state, list) {
			remove(list)
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
		searchList: state => (query, includeArchived = false) => {
			return search(query)
					?.filter(value => value > 0)
					.map(id => state[id])
					.filter(list => list.isArchived === includeArchived)
				|| []
		},
	},
	actions: {
		toggleListFavorite(ctx, list) {
			return ctx.dispatch('updateList', {
				...list,
				isFavorite: !list.isFavorite,
			})
		},

		async createList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			try {
				const createdList = await listService.create(list)
				createdList.namespaceId = list.namespaceId
				ctx.commit('namespaces/addListToNamespace', createdList, {root: true})
				ctx.commit('setList', createdList)
				return createdList
			} finally {
				cancel()
			}
		},

		async updateList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			try {
				await listService.update(list)
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
				return newList
			} catch (e) {
				// Reset the list state to the initial one to avoid confusion for the user
				ctx.commit('setList', {
					...list,
					isFavorite: !list.isFavorite,
				})
				throw e
			} finally {
				cancel()
			}
		},

		async deleteList(ctx, list) {
			const cancel = setLoading(ctx, 'lists')
			const listService = new ListService()

			try {
				const response = await listService.delete(list)
				ctx.commit('removeListById', list)
				ctx.commit('namespaces/removeListFromNamespaceById', list, {root: true})
				removeListFromHistory({id: list.id})
				return response
			} finally {
				cancel()
			}
		},
	},
}