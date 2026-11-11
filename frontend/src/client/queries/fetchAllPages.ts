export type Page<T> = {items?: T[] | null, total_pages?: number}

export async function fetchAllPages<T>(fetchPage: (page: number) => Promise<Page<T>>): Promise<T[]> {
	const items: T[] = []
	for (let page = 1; ; page++) {
		const data = await fetchPage(page)
		items.push(...(data.items ?? []))
		if (page >= (data.total_pages ?? 1)) return items
	}
}
