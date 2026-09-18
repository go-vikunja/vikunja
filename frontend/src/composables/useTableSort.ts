import type {Ref} from 'vue'

export type SortOrder = 'asc' | 'desc'
export type TableSortState<Field extends string> = Partial<Record<Field, SortOrder | 'none'>>

export function useTableSort<Field extends string>(
	sortBy: Ref<TableSortState<Field>>,
	onChange: () => void,
) {
	function ariaSort(field: Field): 'ascending' | 'descending' | undefined {
		const order = sortBy.value[field]
		if (order === 'asc') {
			return 'ascending'
		}
		if (order === 'desc') {
			return 'descending'
		}
		return undefined
	}

	// Allow sorting by multiple columns only when ctrl is pressed
	function sort(field: Field, event?: MouseEvent) {
		const currentOrder = sortBy.value[field]
		let newOrder: SortOrder | undefined
		if (currentOrder === undefined || currentOrder === 'none') {
			newOrder = 'desc'
		} else if (currentOrder === 'desc') {
			newOrder = 'asc'
		}

		const ctrlPressed = event?.ctrlKey || event?.metaKey
		const next: TableSortState<Field> = ctrlPressed ? {...sortBy.value} : {}
		if (newOrder) {
			next[field] = newOrder
		} else {
			delete next[field]
		}
		sortBy.value = next

		onChange()
	}

	return {
		sort,
		ariaSort,
	}
}
