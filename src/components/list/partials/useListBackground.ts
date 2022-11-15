import {ref, watch, type Ref} from 'vue'
import ListService from '@/services/list'
import type {IList} from '@/modelTypes/IList'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'

export function useListBackground(list: Ref<IList>) {
	const background = ref<string | null>(null)
	const backgroundLoading = ref(false)
	const blurHashUrl = ref('')

	watch(
		() => [list.value.id, list.value.backgroundBlurHash] as [IList['id'], IList['backgroundBlurHash']],
		async ([listId, blurHash], oldValue) => {
			if (
				list.value === null ||
				!list.value.backgroundInformation ||
				backgroundLoading.value
			) {
				return
			}

			const [oldListId, oldBlurHash] = oldValue || []
			if (
				oldValue !== undefined && 
				listId === oldListId && blurHash === oldBlurHash
			) {
				// list hasn't changed
				return
			}

			backgroundLoading.value = true

			try {
				const blurHashPromise = getBlobFromBlurHash(blurHash).then((blurHash) => {
					blurHashUrl.value = blurHash ? window.URL.createObjectURL(blurHash) : ''
				})

				const listService = new ListService()
				const backgroundPromise = listService.background(list.value).then((result) => {
					background.value = result
				})
				await Promise.all([blurHashPromise, backgroundPromise])
			} finally {
				backgroundLoading.value = false
			}
		},
		{ immediate: true },
	)

	return {
		background,
		blurHashUrl,
		backgroundLoading,
	}
}