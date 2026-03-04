import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews, createProjects} from '../project/prepareProjects'

test.describe('New Task', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await createProjects(1)
		await BucketFactory.create(1, {
			project_view_id: 4,
		})
	})

	test.describe('New Task Button', () => {
		test('Should show the New Task button on the list view', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/1')
			await expect(page.locator('.switch-view-container .button').filter({hasText: 'New Task'})).toBeVisible()
		})

		test('Should show the New Task button on the table view', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/3')
			await expect(page.locator('.switch-view-container .button').filter({hasText: 'New Task'})).toBeVisible()
		})

		test('Should show the New Task button on the kanban view', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/4')
			await expect(page.locator('.switch-view-container .button').filter({hasText: 'New Task'})).toBeVisible()
		})

		test('Should not show the New Task button for read-only projects', async ({authenticatedPage: page}) => {
			await UserFactory.create(2)
			await UserProjectFactory.create(1, {
				project_id: 2,
				user_id: 1,
				permission: 0,
			})
			const projects = await ProjectFactory.create(2, {
				owner_id: '{increment}',
			})
			await page.goto(`/projects/${projects[1].id}/`)

			await expect(page.locator('.switch-view-container .button').filter({hasText: 'New Task'})).not.toBeVisible()
		})

		test('Should navigate to new task page when clicked', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/1')
			await page.locator('.switch-view-container .button').filter({hasText: 'New Task'}).click()
			await expect(page).toHaveURL(/\/projects\/1\/tasks\/new/)
		})
	})

	test.describe('New Task Detail View', () => {
		test('Should show the new task identifier', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			await expect(page.locator('.task-view h1.title.task-id')).toContainText('New')
		})

		test('Should show the project name', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			await expect(page.locator('.task-view h6.subtitle')).toContainText('First Project')
		})

		test('Should show a Save button', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Save'})).toBeVisible()
		})

		test('Should not show comments section', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			await expect(page.locator('.task-view .comments')).not.toBeVisible()
		})

		test('Should not show the Done button', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			await expect(page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'})).not.toBeVisible()
		})

		test('Should create a task with just a title', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			// Edit the title
			const titleInput = page.locator('.task-view h1.title.input')
			await titleInput.click()
			await titleInput.fill('My New Task')
			await titleInput.blur()

			// Click Save
			const createPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/tasks') && response.request().method() === 'PUT',
			)
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Save'}).click()
			await createPromise

			// Should navigate to the real task detail
			await expect(page).toHaveURL(/\/tasks\/\d+/)
			await expect(page.locator('.task-view h1.title.input')).toContainText('My New Task')
			await expect(page.locator('.global-notification')).toContainText('Success')
		})

		test('Should create a task with a priority', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')

			// Set priority
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Priority'}).click()
			await page.locator('.task-view .columns.details .column').filter({hasText: 'Priority'}).locator('.select select').selectOption('Urgent')

			// Click Save
			const createPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/tasks') && response.request().method() === 'PUT',
			)
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Save'}).click()
			await createPromise

			// Should navigate and retain priority
			await expect(page).toHaveURL(/\/tasks\/\d+/)
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Priority'}).locator('.select select')).toHaveValue('4')
		})

		test('Should create a task with a due date', async ({authenticatedPage: page}) => {
			await page.goto('/projects/1/tasks/new')
			await page.waitForLoadState('networkidle')

			// Set due date via the action button
			const setDueDateButton = page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Due Date'})
			await expect(setDueDateButton).toBeVisible({timeout: 10000})
			await setDueDateButton.click()

			// The due date column should appear with a datepicker
			const dueDateColumn = page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'})
			await expect(dueDateColumn).toBeVisible()

			// Click the datepicker show button to open the calendar
			const datepickerShow = dueDateColumn.locator('.date-input .datepicker .show')
			await expect(datepickerShow).toBeVisible()
			await datepickerShow.click()

			// Click today in the flatpickr calendar (this also closes the datepicker via closeOnChange)
			const todayButton = page.locator('.datepicker-popup .flatpickr-innerContainer .flatpickr-days .flatpickr-day.today')
			await expect(todayButton).toBeVisible()
			await todayButton.click()

			// Click Save
			const createPromise = page.waitForResponse(response =>
				response.url().includes('/projects/1/tasks') && response.request().method() === 'PUT',
			)
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Save'}).click()
			await createPromise

			// Should navigate to the real task detail with due date visible
			await expect(page).toHaveURL(/\/tasks\/\d+/)
			await expect(page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'})).toBeVisible()
		})
	})
})
