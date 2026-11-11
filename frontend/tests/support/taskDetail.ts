import type {Page} from '@playwright/test'

export async function openTaskActions(page: Page) {
	await page.locator('[data-cy="taskDetail.moreActions"]').click()
	return page.locator('.task-view .task-actions-menu .dropdown-menu')
}

export async function clickTaskAction(page: Page, label: string) {
	const menu = await openTaskActions(page)
	await menu.getByRole('button', {name: label, exact: true}).click()
}
