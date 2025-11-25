import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from './prepareProjects'

test.describe('Project View Table', () => {
	// FIXME: Tasks are not loading properly in table view, table shows only headers
	test.skip('Should show a table with tasks', async ({authenticatedPage: page}) => {
		const projects = await createProjects(1)
		const tasks = await TaskFactory.create(1, {
			project_id: 1,
		})
		await page.goto('/projects/1/3')

		await expect(page.locator('.project-table table.table')).toBeVisible()
		await expect(page.locator('.project-table table.table')).toContainText(tasks[0].title)
	})

	test('Should have working column switches', async ({authenticatedPage: page}) => {
		const projects = await createProjects(1)
		await TaskFactory.create(1, {
			project_id: 1,
		})
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)
		await page.goto('/projects/1/3')
		await loadTasksPromise

		// Click the Columns button to open the column selector
		await page.locator('.project-table .filter-container .button').filter({hasText: 'Columns'}).click()

		// Click Priority checkbox to enable Priority column (click on the text like Cypress does)
		await page.locator('.project-table .filter-container .card.columns-filter .card-content').getByText('Priority').click()

		// Wait for Priority checkbox to be checked
		await expect(page.getByRole('checkbox', {name: 'Checkbox Priority'})).toBeChecked()

		// Click Done checkbox to disable Done column (click on the text like Cypress does)
		await page.locator('.project-table .filter-container .card.columns-filter .card-content').getByText('Done', {exact: true}).click()

		// Wait for Done checkbox to be unchecked
		await expect(page.getByRole('checkbox', {name: 'Checkbox Done', exact: true})).not.toBeChecked()

		// Verify Priority column is now visible
		await expect(page.locator('.project-table table.table th').filter({hasText: 'Priority'})).toBeVisible()
		// Verify Done column is now hidden
		await expect(page.locator('.project-table table.table th').filter({hasText: /^Done$/})).not.toBeVisible()
	})

	// FIXME: API returns 500 Internal Server Error when seeding project_views table
	test.skip('Should navigate to the task when the title is clicked', async ({authenticatedPage: page}) => {
		const projects = await createProjects(1)
		const tasks = await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		const loadTasksPromise = page.waitForResponse(response =>
			response.url().includes('/projects/1/views/') && response.url().includes('/tasks'),
		)
		await page.goto('/projects/1/3')
		await loadTasksPromise

		await page.locator('.project-table table.table tbody tr').first().locator('a').first().click()

		await expect(page).toHaveURL(/\/tasks\/\d+/)
	})
})
