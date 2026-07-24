import type {IFilters} from '@/modelTypes/ISavedFilter'
import type {SortBy, Order} from '@/composables/useTaskList'

/**
 * Build a SortBy object from a view's saved filter.sort_by / order_by.
 * Handles both snake_case (API) and camelCase (AbstractModel.assignData) keys.
 * Returns null when the view has no meaningful default sort.
 */
export function sortByFromViewFilter(filter: IFilters | undefined | null): SortBy | null {
	if (!filter) {
		return null
	}

	const raw = filter as IFilters & {
		sortBy?: string[]
		orderBy?: string[]
	}
	const sortBy = raw.sort_by ?? raw.sortBy ?? []
	const orderBy = raw.order_by ?? raw.orderBy ?? []

	if (!Array.isArray(sortBy) || sortBy.length === 0) {
		return null
	}

	const result: SortBy = {}
	for (let i = 0; i < sortBy.length; i++) {
		const field = sortBy[i]
		if (!field || field === 'id') {
			// id is always appended last by formatSortOrder; skip as a sole primary
			continue
		}
		const order = (orderBy[i] === 'desc' ? 'desc' : 'asc') as Order
		;(result as Record<string, Order>)[field] = order
	}

	return Object.keys(result).length > 0 ? result : null
}

/**
 * Serialize a SortBy object into the IFilters sort_by / order_by arrays
 * stored on a project view.
 */
export function viewFilterSortFromSortBy(sortBy: SortBy): Pick<IFilters, 'sort_by' | 'order_by'> {
	const keys = Object.keys(sortBy) as (keyof SortBy)[]
	const sort_by: IFilters['sort_by'] = []
	const order_by: IFilters['order_by'] = []

	for (const key of keys) {
		const order = sortBy[key]
		if (!order || order === 'none') {
			continue
		}
		sort_by.push(key as IFilters['sort_by'][number])
		order_by.push(order)
	}

	return {sort_by, order_by}
}

/**
 * Encode a single primary sort selection as used by SortPopup / ViewEditForm
 * (`field:order`, with `position:asc` meaning manual).
 */
export function encodeSortSelection(sortBy: SortBy | null | undefined, manual = 'position:asc'): string {
	if (!sortBy) {
		return manual
	}
	const key = Object.keys(sortBy)[0]
	if (!key || key === 'position') {
		return manual
	}
	const order = (sortBy as Record<string, Order>)[key] ?? 'asc'
	return `${key}:${order}`
}

export function decodeSortSelection(value: string): SortBy {
	const [field, order] = value.split(':') as [string, Order]
	const sort: SortBy = {}
	;(sort as Record<string, Order>)[field || 'position'] = order === 'desc' ? 'desc' : 'asc'
	return sort
}
