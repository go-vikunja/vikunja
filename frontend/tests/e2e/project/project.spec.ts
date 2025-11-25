import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {createProjects} from './prepareProjects'

test.describe('Projects', () => {
	test.use({
		// Use authenticated page for all tests
	})

	let projects: any[]

	test.beforeEach(async ({authenticatedPage}) => {
		projects = await createProjects()
	})

	test('Should create a new project', async ({authenticatedPage: page}) => {
		await page.goto('/projects')
		await page.waitForLoadState('networkidle')
		await page.locator('.action-buttons').getByRole('link', {name: /project/i}).click()
		await expect(page).toHaveURL(/\/projects\/new/)
		await expect(page.locator('.card-header-title')).toContainText('New project')
		await page.locator('input[name=projectTitle]').fill('New Project')
		await page.locator('.button').filter({hasText: 'Create'}).click()

		await expect(page.locator('.global-notification', {timeout: 1000})).toContainText('Success')
		await expect(page).toHaveURL(/\/projects\//)
		await expect(page.locator('.project-title')).toContainText('New Project')
	})

	// FIXME: API returns 500 Internal Server Error when seeding project_views table
	test.skip('Should redirect to a specific project view after visited', async ({authenticatedPage: page}) => {
		const loadBucketsPromise = page.waitForResponse(response =>
			response.url().includes('/projects/') &&
			response.url().includes('/views/') &&
			response.url().includes('/tasks'),
		)

		await page.goto('/projects/1/4')
		await expect(page).toHaveURL(/\/projects\/1\/4/)
		await loadBucketsPromise
		await page.goto('/projects/1')
		await expect(page).toHaveURL(/\/projects\/1\/4/)
	})

	// FIXME: seeding fails with error 500
	test.skip('Should rename the project in all places', async ({authenticatedPage: page}) => {
		await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		const newProjectName = 'New project name'

		// Navigate to project and wait for redirect to view
		await page.goto('/projects/1/1')
		await page.waitForLoadState('networkidle')
		await expect(page.locator('.project-title')).toContainText('First Project')

		// Click the project title dropdown and select Edit
		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^edit$/i}).click()
		await page.waitForLoadState('networkidle')

		// Fill in the new name
		await page.locator('input#title').fill(newProjectName)
		await page.locator('footer.card-footer .button').filter({hasText: /^Save$/}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.project-title')).toContainText(newProjectName)
		await expect(page.locator('.project-title')).not.toContainText(projects[0].title)
		await expect(page.locator('.menu-container .menu-list').getByRole('listitem').filter({hasText: newProjectName})).toBeVisible()
		await page.goto('/')
		await expect(page.locator('.project-grid')).toContainText(newProjectName)
		await expect(page.locator('.project-grid')).not.toContainText(projects[0].title)
	})

	test('Should remove a project when deleting it', async ({authenticatedPage: page}) => {
		await page.goto(`/projects/${projects[0].id}/1`)
		await page.waitForLoadState('networkidle')

		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^delete$/i}).click()
		await page.waitForLoadState('networkidle')

		await expect(page).toHaveURL(/\/settings\/delete/)
		await page.getByRole('button', {name: /do it/i}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page).toHaveURL('/')
		await expect(page.getByRole('link', {name: projects[0].title})).not.toBeVisible()
	})

	test('Should archive a project', async ({authenticatedPage: page}) => {
		await page.goto(`/projects/${projects[0].id}/1`)
		await page.waitForLoadState('networkidle')

		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^archive$/i}).click()
		await expect(page.locator('.modal-content')).toContainText('Archive this project')
		await page.getByRole('button', {name: /do it/i}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('main.app-content')).toContainText('This project is archived. It is not possible to create new or edit tasks for it.')
	})

	// FIXME: API returns 500 Internal Server Error when seeding project_views table
	test.skip('Should show all projects on the projects page', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(10)

		await page.goto('/projects')
		await page.waitForLoadState('networkidle')

		for (const p of projects) {
			await expect(page.locator('.project-grid')).toContainText(p.title)
		}
	})

	// FIXME: API returns 500 Internal Server Error when seeding project_views table
	test.skip('Should not show archived projects if the filter is not checked', async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1, {
			id: 2,
		}, false)
		await ProjectFactory.create(1, {
			id: 3,
			is_archived: true,
		}, false)

		// Initial
		await page.goto('/projects')
		await page.waitForLoadState('networkidle')
		await expect(page.locator('.project-grid')).not.toContainText('Archived')

		// Show archived - click the checkbox label text
		await page.getByText('Show Archived').click()
		await expect(page.locator('input[type="checkbox"]').first()).toBeChecked()
		await expect(page.locator('.project-grid')).toContainText('Archived')

		// Don't show archived
		await page.getByText('Show Archived').click()
		await expect(page.locator('input[type="checkbox"]').first()).not.toBeChecked()

		// Second time visiting after unchecking
		await page.goto('/projects')
		await page.waitForLoadState('networkidle')
		await expect(page.locator('input[type="checkbox"]').first()).not.toBeChecked()
		await expect(page.locator('.project-grid')).not.toContainText('Archived')
	})
})
