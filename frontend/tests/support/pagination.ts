import type {APIRequestContext} from '@playwright/test'

export async function serverPageSize(apiContext: APIRequestContext): Promise<number> {
	const response = await apiContext.get('info')
	return (await response.json()).max_items_per_page
}
