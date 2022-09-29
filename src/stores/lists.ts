import {watch, reactive, shallowReactive, unref, toRefs, readonly} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {useI18n} from 'vue-i18n'

import ListService from '@/services/list'
import {setLoadingPinia} from '@/store/helper'
import {removeListFromHistory} from '@/modules/listHistory'
import {createNewIndexer} from '@/indexes'
import {useNamespaceStore} from './namespaces'

import type {ListState} from '@/store/types'
import type {IList} from '@/modelTypes/IList'

import type {MaybeRef} from '@vueuse/core'

import ListModel from '@/models/list'
import {success} from '@/message'

const {add, remove, search, update} = createNewIndexer('lists', ['title', 'description'])

const FavoriteListsNamespace = -2

export const useListStore = defineStore('list', {
	state: () : ListState => ({
		isLoading: false,
		// The lists are stored as an object which has the list ids as keys.
		lists: {},
	}),

	getters: {
		getListById(state) {
			return (id: IList['id']) => typeof state.lists[id] !== 'undefined' ? state.lists[id] : null
		},

		findListByExactname(state) {
			return (name: string) => {
				const list = Object.values(state.lists).find(l => {
					return l.title.toLowerCase() === name.toLowerCase()
				})
				return typeof list === 'undefined' ? null : list
			}
		},

		searchList(state) {
			return (query: string, includeArchived = false) => {
				return search(query)
					?.filter(value => value > 0)
					.map(id => state.lists[id])
					.filter(list => list.isArchived === includeArchived)
					|| []
			}
		},
	},

	actions: {
		setIsLoading(isLoading: boolean) {
			this.isLoading = isLoading
		},

		setList(list: IList) {
			this.lists[list.id] = list
			update(list)
		},

		setLists(lists: IList[]) {
			lists.forEach(l => {
				this.lists[l.id] = l
				add(l)
			})
		},

		removeListById(list: IList) {
			remove(list)
			delete this.lists[list.id]
		},

		toggleListFavorite(list: IList) {
			// The favorites pseudo list is always favorite
			// Archived lists cannot be marked favorite
			if (list.id === -1 || list.isArchived) {
				return
			}
			return this.updateList({
				...list,
				isFavorite: !list.isFavorite,
			})
		},

		async createList(list: IList) {
			const cancel = setLoadingPinia(this)
			const listService = new ListService()

			try {
				const createdList = await listService.create(list)
				createdList.namespaceId = list.namespaceId
				const namespaceStore = useNamespaceStore()
				namespaceStore.addListToNamespace(createdList)
				this.setList(createdList)
				return createdList
			} finally {
				cancel()
			}
		},

		async updateList(list: IList) {
			const cancel = setLoadingPinia(this)
			const listService = new ListService()

			try {
				await listService.update(list)
				this.setList(list)
				const namespaceStore = useNamespaceStore()
				namespaceStore.setListInNamespaceById(list)

				// the returned list from listService.update is the same!
				// in order to not validate vuex mutations we have to create a new copy
				const newList = {
					...list,
					namespaceId: FavoriteListsNamespace,
				}
				if (list.isFavorite) {
					namespaceStore.addListToNamespace(newList)
				} else {
					namespaceStore.removeListFromNamespaceById(newList)
				}
				namespaceStore.loadNamespacesIfFavoritesDontExist(null)
				namespaceStore.removeFavoritesNamespaceIfEmpty(null)
				return newList
			} catch (e) {
				// Reset the list state to the initial one to avoid confusion for the user
				this.setList({
					...list,
					isFavorite: !list.isFavorite,
				})
				throw e
			} finally {
				cancel()
			}
		},

		async deleteList(list: IList) {
			const cancel = setLoadingPinia(this)
			const listService = new ListService()

			try {
				const response = await listService.delete(list)
				this.removeListById(list)
				const namespaceStore = useNamespaceStore()
				namespaceStore.removeListFromNamespaceById(list)
				removeListFromHistory({id: list.id})
				return response
			} finally {
				cancel()
			}
		},
	},
})

export function useList(listId: MaybeRef<IList['id']>) {
	const listService = shallowReactive(new ListService())
	const {loading: isLoading} = toRefs(listService)
	const list: ListModel = reactive(new ListModel())
	const {t} = useI18n({useScope: 'global'})

	watch(
		() => unref(listId),
		async (listId) => {
			const loadedList = await listService.get(new ListModel({id: listId}))
			Object.assign(list, loadedList)
		},
		{immediate: true},
	)

	const listStore = useListStore()
	async function save() {
		await listStore.updateList(list)
		success({message: t('list.edit.success')})
	}

	return {
		isLoading: readonly(isLoading),
		list,
		save,
	}
}

// support hot reloading
if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useListStore, import.meta.hot))
}