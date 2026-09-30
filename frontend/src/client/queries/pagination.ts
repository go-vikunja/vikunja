import {
	hashKey,
	type QueryClient,
	type QueryKey,
} from '@tanstack/vue-query'

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

type PaginatedResponse<T> = {
	items?: T[] | null,
	page?: number,
	per_page?: number,
	total?: number,
	total_pages?: number,
}

export function toPaginated<T>(data: PaginatedResponse<T>, page: number): Paginated<T> {
	return {
		items: data.items ?? [],
		page: data.page ?? page,
		per_page: data.per_page ?? 0,
		total: data.total ?? 0,
		total_pages: data.total_pages ?? 0,
	}
}

export function normalizePageNumber(page: unknown): number {
	const parsed = Number(page)
	return Number.isInteger(parsed) && parsed >= 1 ? parsed : 1
}

// Keys end in the page number; only scopes (key minus page) that held the item lose one from total.
export function removeFromPages<T>(client: QueryClient, queryKey: QueryKey, matches: (item: T) => boolean) {
	const cached = client.getQueriesData<Paginated<T>>({queryKey})
	const scopeOf = (key: QueryKey) => hashKey(key.slice(0, -1))
	const hitScopes = new Set(cached
		.filter(([, page]) => page?.items.some(matches))
		.map(([key]) => scopeOf(key)))
	for (const [key, page] of cached) {
		if (!page || !hitScopes.has(scopeOf(key))) continue
		const total = Math.max(0, page.total - 1)
		client.setQueryData<Paginated<T>>(key, {
			...page,
			items: page.items.filter(item => !matches(item)),
			total,
			total_pages: totalPagesFor(page, total),
		})
	}
}
