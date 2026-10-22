import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {createDefaultViews, createProjects} from './prepareProjects'
import {createTasksWithPriorities, createTasksWithSearch} from '../../support/filterTestHelpers'

test.describe('Project View Table', () => {
	test('Should show a table with tasks', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const tasks = await TaskFactory.create(1, {
			project_id: 1,
		})
		await page.goto('/projects/1/3')

		await expect(page.locator('.project-table table.table')).toBeVisible()
		await expect(page.locator('.project-table table.table')).toContainText(tasks[0].title)
	})

	test('Should have working column switches', async ({authenticatedPage: page}) => {
		await createProjects(1)
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

	test('Should show every priority for open and completed tasks', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const priorities = ['Unset', 'Low', 'Medium', 'High', 'Urgent', 'DO NOW']
		const tasks = await TaskFactory.create(priorities.length * 2, {
			project_id: 1,
			title: (i: number) => `Priority task ${i}`,
			priority: (i: number) => (i - 1) % priorities.length,
			done: (i: number) => i > priorities.length,
		})
		await page.goto('/projects/1/3')
		await page.getByRole('button', {name: 'Columns', exact: true}).click()
		await page.locator('.columns-filter').getByText('Priority', {exact: true}).click()
		await page.getByRole('button', {name: 'Columns', exact: true}).click()

		for (const task of tasks) {
			const row = page.locator('.project-table tbody tr').filter({
				has: page.getByRole('link', {name: task.title, exact: true}),
			})
			await expect(row).toBeVisible()
			await expect(row.locator('.priority-label')).toHaveText(priorities[task.priority])
		}
	})

	test('Should navigate to the task when the title is clicked', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await TaskFactory.create(5, {
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

	test('Should respect filter query parameter from URL', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const {highPriorityTasks, lowPriorityTasks} = await createTasksWithPriorities()

		await page.goto('/projects/1/3?filter=priority%20>=%204')

		await expect(page).toHaveURL(/filter=priority/)

		// Wait for tasks to load and verify high priority tasks are visible
		await expect(page.locator('.project-table table.table')).toContainText(highPriorityTasks[0].title, {timeout: 10000})
		await expect(page.locator('.project-table table.table')).toContainText(highPriorityTasks[1].title)

		// Verify low priority tasks are not visible
		await expect(page.locator('.project-table table.table')).not.toContainText(lowPriorityTasks[0].title)
		await expect(page.locator('.project-table table.table')).not.toContainText(lowPriorityTasks[1].title)
	})

	test('Should respect search query parameter from URL', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const {searchableTask} = await createTasksWithSearch()

		await page.goto('/projects/1/3?s=meeting')

		await expect(page).toHaveURL(/s=meeting/)

		// Wait for search results to load and verify searchable task is visible
		await expect(page.locator('.project-table table.table')).toContainText(searchableTask.title, {timeout: 10000})

		// Verify only one task row is shown (the search result)
		await expect(page.locator('.project-table table.table tbody tr')).toHaveCount(1)
	})

	test('Should create a task from the table view', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await TaskFactory.create(1, {
			project_id: 1,
		})
		await page.goto('/projects/1/3')
		await expect(page.locator('.project-table tbody tr')).toHaveCount(1)

		await page.locator('.project-table .task-add textarea').fill('Task from the table')
		await page.locator('.project-table .task-add textarea').press('Enter')

		await expect(page.locator('.project-table tbody tr')).toHaveCount(2)
		await expect(page.locator('.project-table table.table')).toContainText('Task from the table')
	})

	test('Should not show the task input for an archived project', async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1, {
			is_archived: true,
		})
		await createDefaultViews(1)
		await TaskFactory.create(1, {
			project_id: 1,
		})
		await page.goto('/projects/1/3')

		await expect(page.locator('.project-table tbody tr')).toHaveCount(1)
		await expect(page.locator('.project-table .task-add')).toHaveCount(0)
	})

	test('Should not show the task input for a project shared read only', async ({authenticatedPage: page}) => {
		await UserFactory.create(2)
		await ProjectFactory.create(2, {
			owner_id: '{increment}',
		})
		await UserProjectFactory.create(1, {
			project_id: 2,
			user_id: 1,
			permission: 0,
		})
		await createDefaultViews(2)
		await TaskFactory.create(1, {
			project_id: 2,
		})
		await page.goto('/projects/2/3')

		await expect(page.locator('.project-table tbody tr')).toHaveCount(1)
		await expect(page.locator('.project-table .task-add')).toHaveCount(0)
	})
})
