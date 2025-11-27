import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from './prepareProjects'

async function openAndSetFilters(page) {
	await page.locator('.filter-container button').filter({hasText: 'Filters'}).click()
	await expect(page.locator('.filter-popup')).toBeVisible()
	await page.locator('.filter-popup .filter-input .ProseMirror').fill('done = true')
	await page.locator('.filter-popup button').filter({hasText: 'Show results'}).click()
}

test.describe('Filter Persistence Across Views', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await createProjects()
		await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
			title: 'Test Task {increment}',
		})
		await page.goto('/projects/1/1')
	})

	test('should persist filters in List view after page refresh', async ({authenticatedPage: page}) => {
		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should persist filters in Table view after page refresh', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/3')

		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should persist filters in Kanban view after page refresh', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/4')

		await openAndSetFilters(page)

		await expect(page).toHaveURL(/filter=/)

		await page.reload()

		await expect(page).toHaveURL(/filter=/)
	})

	test('should handle URL sharing with filters', async ({authenticatedPage: page}) => {
		// Visit URL with pre-existing filter parameters
		await page.goto('/projects/1/4?filter=done%3Dtrue&s=Test')

		// Verify URL parameters are preserved
		await expect(page).toHaveURL(/filter=done%3Dtrue/)
		await expect(page).toHaveURL(/s=Test/)

		// Switch views and verify parameters persist
		await page.goto('/projects/1/3?filter=done%3Dtrue&s=Test')
		await expect(page).toHaveURL(/filter=done%3Dtrue/)
		await expect(page).toHaveURL(/s=Test/)
	})
})
