import {watch, reactive, shallowReactive, unref, toRefs, readonly, ref, computed} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'
import {useI18n} from 'vue-i18n'

import ListService from '@/services/list'
import {setModuleLoading} from '@/stores/helper'
import {removeListFromHistory} from '@/modules/listHistory'
import {createNewIndexer} from '@/indexes'
import {useNamespaceStore} from './namespaces'

import type {IList} from '@/modelTypes/IList'

import type {MaybeRef} from '@vueuse/core'

import ListModel from '@/models/list'
import {success} from '@/message'
import {useBaseStore} from '@/stores/base'

const {add, remove, search, update} = createNewIndexer('lists', ['title', 'description'])

const FavoriteListsNamespace = -2

export interface ListState {
	[id: IList['id']]: IList
}

export const useListStore = defineStore('list', () => {
	const baseStore = useBaseStore()
	const namespaceStore = useNamespaceStore()

	const isLoading = ref(false)

	// The lists are stored as an object which has the list ids as keys.
	const lists = ref<ListState>({})


	const getListById = computed(() => {
		return (id: IList['id']) => typeof lists.value[id] !== 'undefined' ? lists.value[id] : null
	})

	const findListByExactname = computed(() => {
		return (name: string) => {
			const list = Object.values(lists.value).find(l => {
				return l.title.toLowerCase() === name.toLowerCase()
			})
			return typeof list === 'undefined' ? null : list
		}
	})

	const searchList = computed(() => {
		return (query: string, includeArchived = false) => {
			return search(query)
				?.filter(value => value > 0)
				.map(id => lists.value[id])
				.filter(list => list.isArchived === includeArchived)
				|| []
		}
	})

	function setIsLoading(newIsLoading: boolean) {
		isLoading.value = newIsLoading
	}

	function setList(list: IList) {
		lists.value[list.id] = list
		update(list)

		if (baseStore.currentList?.id === list.id) {
			baseStore.setCurrentList(list)
		}
	}

	function setLists(newLists: IList[]) {
		newLists.forEach(l => {
			lists.value[l.id] = l
			add(l)
		})
	}

	function removeListById(list: IList) {
		remove(list)
		delete lists.value[list.id]
	}

	function toggleListFavorite(list: IList) {
		// The favorites pseudo list is always favorite
		// Archived lists cannot be marked favorite
		if (list.id === -1 || list.isArchived) {
			return
		}
		return updateList({
			...list,
			isFavorite: !list.isFavorite,
		})
	}

	async function createList(list: IList) {
		const cancel = setModuleLoading(setIsLoading)
		const listService = new ListService()

		try {
			const createdList = await listService.create(list)
			createdList.namespaceId = list.namespaceId
			namespaceStore.addListToNamespace(createdList)
			setList(createdList)
			return createdList
		} finally {
			cancel()
		}
	}

	async function updateList(list: IList) {
		const cancel = setModuleLoading(setIsLoading)
		const listService = new ListService()

		try {
			await listService.update(list)
			setList(list)
			namespaceStore.setListInNamespaceById(list)

			// the returned list from listService.update is the same!
			// in order to not create a manipulation in pinia store we have to create a new copy
			const newList = {
				...list,
				namespaceId: FavoriteListsNamespace,
			}
			if (list.isFavorite) {
				namespaceStore.addListToNamespace(newList)
			} else {
				namespaceStore.removeListFromNamespaceById(newList)
			}
			namespaceStore.loadNamespacesIfFavoritesDontExist()
			namespaceStore.removeFavoritesNamespaceIfEmpty()
			return newList
		} catch (e) {
			// Reset the list state to the initial one to avoid confusion for the user
			setList({
				...list,
				isFavorite: !list.isFavorite,
			})
			throw e
		} finally {
			cancel()
		}
	}

	async function deleteList(list: IList) {
		const cancel = setModuleLoading(setIsLoading)
		const listService = new ListService()

		try {
			const response = await listService.delete(list)
			removeListById(list)
			namespaceStore.removeListFromNamespaceById(list)
			removeListFromHistory({id: list.id})
			return response
		} finally {
			cancel()
		}
	}

	return {
		isLoading: readonly(isLoading),
		lists: readonly(lists),

		getListById,
		findListByExactname,
		searchList,

		setList,
		setLists,
		removeListById,
		toggleListFavorite,
		createList,
		updateList,
		deleteList,
	}
})

export function useList(listId: MaybeRef<IList['id']>) {
	const listService = shallowReactive(new ListService())
	const {loading: isLoading} = toRefs(listService)
	const list: IList = reactive(new ListModel())
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