import {computed, ref, shallowRef, watch, type Ref, type ShallowRef} from 'vue'
import type {UseQueryReturnType} from '@tanstack/vue-query'

import {clampPage, type Paginated} from '@/client/queries/pagination'
import {captureFocus} from '@/composables/useFocusAfterRemoval'
import {usePendingRows} from '@/composables/usePendingRows'

type RowMutation = {mutateAsync: (id: number) => Promise<unknown>}
// Matches data-trash-action in TrashRowActions.vue
type RowAction = 'restore' | 'delete-forever'
type MoveFocus = ReturnType<typeof captureFocus>

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
	// Focus is captured on the click that opened the modal, not on its submit button
	const deleteModal: ShallowRef<{item: T, moveFocus: MoveFocus} | null> = shallowRef(null)
	const emptyModal = shallowRef<{moveFocus: MoveFocus} | null>(null)
	const itemToDelete = computed(() => deleteModal.value?.item ?? null)
	const showEmptyModal = computed(() => emptyModal.value !== null)

	const listElement = ref<HTMLElement | null>(null)
	const emptyStateElement = ref<HTMLElement | null>(null)

	function rowAction(id: number, action: RowAction) {
		return listElement.value?.querySelector<HTMLElement>(
			`[data-trash-row="${id}"] [data-trash-action="${action}"]`,
		)
	}

	// The next row's same action, else the previous row's, else the list itself
	async function removeRow(mutation: RowMutation, id: number, action: RowAction, moveFocus: MoveFocus) {
		const ids = items.value.map(item => item.id)
		const index = ids.indexOf(id)
		if (!await pendingRows.run(mutation, id)) {
			return
		}
		const neighbours = [
			...ids.slice(index + 1),
			...ids.slice(0, index).reverse(),
		]
		await moveFocus(
			// Not every row offers delete forever
			...neighbours.map(neighbour => () => rowAction(neighbour, action) ?? rowAction(neighbour, 'restore')),
			() => emptyStateElement.value,
			() => listElement.value,
		)
	}

	function restore(id: number) {
		return removeRow(options.restore, id, 'restore', captureFocus())
	}

	function confirmDeleteForever(item: T) {
		if (pendingRows.has(item.id)) {
			return
		}
		deleteModal.value = {
			item,
			moveFocus: captureFocus(),
		}
	}

	function cancelDeleteForever() {
		deleteModal.value = null
	}

	async function deleteForever() {
		const modal = deleteModal.value
		deleteModal.value = null
		if (!modal) {
			return
		}
		await removeRow(options.deleteForever, modal.item.id, 'delete-forever', modal.moveFocus)
	}

	function openEmptyModal() {
		if (total.value === 0) {
			return
		}
		emptyModal.value = {moveFocus: captureFocus()}
	}

	function closeEmptyModal() {
		emptyModal.value = null
	}

	async function emptyTrash() {
		const modal = emptyModal.value
		emptyModal.value = null
		if (!modal) {
			return
		}
		try {
			await options.empty.mutateAsync()
		} catch {
			return
		}
		currentPage.value = 1
		await modal.moveFocus(
			() => emptyStateElement.value,
			() => listElement.value,
		)
	}

	return {
		currentPage,
		items,
		total,
		totalPages,
		isFetching: query.isFetching,
		showEmptyState,
		listElement,
		emptyStateElement,
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
