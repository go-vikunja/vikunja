import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'

/**
 * Tests for filter autocomplete functionality, specifically for:
 * - Project names with spaces (Issue #2010)
 * - Verifying filters save correctly without corruption
 */

async function createProjectWithViews(id: number, title: string, ownerId: number, truncate = false) {
	await ProjectFactory.create(1, {
		id,
		title,
		owner_id: ownerId,
	}, truncate)
	await ProjectViewFactory.create(1, {
		id: id * 4,
		project_id: id,
		view_kind: 0, // List
	}, false)
}

// Helper to get the filter input ProseMirror editor
function getFilterInput(page) {
	return page.locator('.filter-input .ProseMirror')
}

test.describe('Filter Autocomplete', () => {
	test.beforeEach(async ({authenticatedPage, currentUser}) => {
		// authenticatedPage fixture triggers apiContext which sets up Factory.request
		await ProjectFactory.truncate()
		await TaskFactory.truncate()
		await ProjectViewFactory.truncate()

		const userId = currentUser.id

		// Create projects - one with spaces in name (the bug case)
		await createProjectWithViews(1, 'Inbox', userId)
		await createProjectWithViews(2, 'Work To Do', userId)
		await createProjectWithViews(3, 'Personal Tasks', userId)

		// Create tasks in each project
		await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			title: 'Inbox Task',
		})
		await TaskFactory.create(1, {
			id: 2,
			project_id: 2,
			title: 'Work Task 1',
		})
		await TaskFactory.create(1, {
			id: 3,
			project_id: 2,
			title: 'Work Task 2',
		})
	})

	test.describe('Saved Filter Creation with Autocomplete', () => {
		test('should replace single-word project name via autocomplete', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			// Wait for projects to be loaded
			await expect(page.getByRole('link', {name: 'Inbox', exact: true})).toBeVisible({timeout: 10000})

			// Fill in filter name
			await page.locator('input#Title').fill('Inbox Filter')

			// Type filter with project name to trigger autocomplete
			const filterInput = getFilterInput(page)
			await filterInput.click()
			await page.keyboard.press('ControlOrMeta+a')
			await page.keyboard.press('Backspace')
			await filterInput.pressSequentially('project = Inb', {delay: 50})

			// Wait for autocomplete popup and select "Inbox"
			const autocompletePopup = page.locator('#filter-autocomplete-popup')
			await expect(autocompletePopup).toBeVisible({timeout: 5000})
			await autocompletePopup.getByRole('button', {name: 'Inbox'}).click()

			// Verify the filter text is correct after autocomplete replacement
			await expect(filterInput).toContainText('project = Inbox')

			// Save the filter and verify no error
			await page.locator('button.is-primary.is-fullwidth').click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
		})

		test('should replace multi-word project name with spaces via autocomplete', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			// Wait for projects to be loaded
			await expect(page.locator('.menu-list a').filter({hasText: 'Work To Do'})).toBeVisible({timeout: 10000})

			// Fill in filter name
			await page.locator('input#Title').fill('Work Filter')

			// Type filter with partial project name to trigger autocomplete
			const filterInput = getFilterInput(page)
			await filterInput.click()
			await page.keyboard.press('ControlOrMeta+a')
			await page.keyboard.press('Backspace')
			await filterInput.pressSequentially('project = Work', {delay: 50})

			// Wait for autocomplete popup to update with latest context
			const autocompletePopup = page.locator('#filter-autocomplete-popup')
			await expect(autocompletePopup).toBeVisible({timeout: 5000})
			// Wait a bit for the context to be fully updated after the last keystroke
			await page.waitForTimeout(100)
			await autocompletePopup.getByRole('button', {name: 'Work To Do'}).click()

			// Verify the filter text is correct after autocomplete replacement
			// Multi-word names should be quoted
			await expect(filterInput).toContainText('project = "Work To Do"')

			// Save the filter and verify no error
			await page.locator('button.is-primary.is-fullwidth').click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
		})

		test('should handle autocomplete after logical operator', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			// Wait for projects to be loaded
			await expect(page.getByRole('link', {name: 'Inbox', exact: true})).toBeVisible({timeout: 10000})

			await page.locator('input#Title').fill('Complex Filter')

			const filterInput = getFilterInput(page)
			await filterInput.click()
			await page.keyboard.press('ControlOrMeta+a')
			await page.keyboard.press('Backspace')

			// Type a complex filter with autocomplete for project name
			await filterInput.pressSequentially('done = false && project = Pers', {delay: 50})

			// Wait for autocomplete popup and select "Personal Tasks"
			const autocompletePopup = page.locator('#filter-autocomplete-popup')
			await expect(autocompletePopup).toBeVisible({timeout: 5000})
			await autocompletePopup.getByRole('button', {name: 'Personal Tasks'}).click()

			// Verify correct filter text - multi-word name should be quoted
			await expect(filterInput).toContainText('done = false && project = "Personal Tasks"')

			// Save and verify no error
			await page.locator('button.is-primary.is-fullwidth').click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
		})
	})
})
