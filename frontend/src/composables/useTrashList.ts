import {computed, ref, shallowRef, watch, type Ref, type ShallowRef} from 'vue'
import type {UseQueryReturnType} from '@tanstack/vue-query'

import {clampPage, type Paginated} from '@/client/queries/pagination'
import {usePendingRows} from '@/composables/usePendingRows'

type RowMutation = {mutateAsync: (id: number) => Promise<unknown>}

export function useTrashList<T extends {id: number}>(options: {
	// A callback, so the caller's useQuery keeps the concrete types
	query: (page: Ref<number>) => UseQueryReturnType<Paginated<T>, Error>,
	restore: RowMutation,
	deleteForever: RowMutation,
	empty: {mutateAsync: () => Promise<unknown>},
}) {
	const currentPage = ref(1)
	const query = options.query(currentPage)
	const items = computed(() => query.data.value?.items ?? [])
	const total = computed(() => query.data.value?.total ?? 0)
	const totalPages = computed(() => query.data.value?.total_pages ?? 0)
	const showEmptyState = computed(() => items.value.length === 0
		&& query.isSuccess.value
		&& !query.isFetching.value)

	// Removing the last row of the last page leaves the page empty
	watch(
		() => clampPage(currentPage.value, {
			data: query.data.value,
			isPlaceholderData: query.isPlaceholderData.value,
		}),
		page => {
			currentPage.value = page
		},
	)

	const pendingRows = usePendingRows()
	const itemToDelete: ShallowRef<T | null> = shallowRef(null)
	const showEmptyModal = ref(false)

	function restore(id: number) {
		return pendingRows.run(options.restore, id)
	}

	function confirmDeleteForever(item: T) {
		if (pendingRows.has(item.id)) {
			return
		}
		itemToDelete.value = item
	}

	function cancelDeleteForever() {
		itemToDelete.value = null
	}

	async function deleteForever() {
		const item = itemToDelete.value
		itemToDelete.value = null
		if (!item) {
			return
		}
		await pendingRows.run(options.deleteForever, item.id)
	}

	function openEmptyModal() {
		if (total.value === 0) {
			return
		}
		showEmptyModal.value = true
	}

	function closeEmptyModal() {
		showEmptyModal.value = false
	}

	async function emptyTrash() {
		showEmptyModal.value = false
		try {
			await options.empty.mutateAsync()
		} catch {
			return
		}
		currentPage.value = 1
	}

	return {
		currentPage,
		items,
		total,
		totalPages,
		isFetching: query.isFetching,
		showEmptyState,
		isPending: pendingRows.has,
		itemToDelete,
		showEmptyModal,
		restore,
		confirmDeleteForever,
		cancelDeleteForever,
		deleteForever,
		openEmptyModal,
		closeEmptyModal,
		emptyTrash,
	}
}
