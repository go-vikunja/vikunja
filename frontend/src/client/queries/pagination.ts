export type Paginated<T> = {
	items: T[],
	page: number,
	per_page: number,
	total: number,
	total_pages: number,
}

// The v2 schema maximum; a larger per_page is rejected with 422.
export const API_MAX_PER_PAGE = 1000

export function totalPagesFor(
	page: Pick<Paginated<unknown>, 'per_page' | 'total_pages'>,
	total: number,
): number {
	return page.per_page > 0 ? Math.ceil(total / page.per_page) : page.total_pages
}

export function pageSizeFor(configured: number): number {
	return Math.min(Math.max(configured, 1), API_MAX_PER_PAGE)
}
