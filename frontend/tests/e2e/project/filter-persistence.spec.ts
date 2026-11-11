import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from './prepareProjects'

async function openFilterPopup(page) {
	await page.locator('.filter-container button').filter({hasText: 'Filters'}).click()
	await expect(page.locator('.filter-popup')).toBeVisible()
}

async function openAndSetFilters(page, filter = 'done = true') {
	await openFilterPopup(page)
	await page.locator('.filter-popup .filter-input .ProseMirror').fill(filter)
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

	test('should only show a task without a due date when include nulls is ticked', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			title: 'Overdue Task',
			due_date: new Date(Date.now() - 86_400_000).toISOString(),
		})
		await TaskFactory.create(1, {
			id: 2,
			project_id: 1,
			title: 'No Due Date Task',
		}, false)

		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks')).toContainText('No Due Date Task')

		await openAndSetFilters(page, 'due_date < now')

		await expect(page.locator('.tasks')).toContainText('Overdue Task')
		await expect(page.locator('.tasks')).not.toContainText('No Due Date Task')

		const requestWithNulls = page.waitForRequest(request => {
			const url = new URL(request.url())
			return url.pathname.endsWith('/projects/1/views/1/tasks') &&
				url.searchParams.get('filter_include_nulls') === 'true'
		})

		await openFilterPopup(page)
		const includeNulls = page.locator('.filter-popup .fancy-checkbox').filter({hasText: 'Include Tasks which'})
		// The checkbox stretches across the flex column, so its centre misses the label.
		await includeNulls.locator('label').click()
		await expect(includeNulls.locator('input[type=checkbox]')).toBeChecked()
		await page.locator('.filter-popup button').filter({hasText: 'Show results'}).click()
		await requestWithNulls

		await expect(page.locator('.tasks')).toContainText('No Due Date Task')
		await expect(page.locator('.tasks')).toContainText('Overdue Task')
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
