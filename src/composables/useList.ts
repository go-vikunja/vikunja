import {watch, reactive, shallowReactive, unref, toRefs, readonly} from 'vue'
import type {MaybeRef} from '@vueuse/core'
import {useStore} from '@/store'

import ListService from '@/services/list'
import ListModel from '@/models/list'
import { success } from '@/message'
import {useI18n} from 'vue-i18n'

export function useList(listId: MaybeRef<ListModel['id']>) {
	const listService = shallowReactive(new ListService())
	const {loading: isLoading} = toRefs(listService)
	const list : ListModel = reactive(new ListModel({}))
	const {t} = useI18n({useScope: 'global'})

	watch(
		() => unref(listId),
		async (listId) => {
			const loadedList = await listService.get(new ListModel({id: listId}))
			Object.assign(list, loadedList)
		},
		{immediate: true},
	)


	const store = useStore()
	async function save() {
		await store.dispatch('lists/updateList', list)
		success({message: t('list.edit.success')})
	}

	return {
		isLoading: readonly(isLoading),
		list,
		save,
	}
}