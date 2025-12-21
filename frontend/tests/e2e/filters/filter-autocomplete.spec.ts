import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {SavedFilterFactory} from '../../factories/saved_filter'

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
		await SavedFilterFactory.truncate()

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

			// Wait for autocomplete popup with the specific option we want to click
			const autocompletePopup = page.locator('#filter-autocomplete-popup')
			await expect(autocompletePopup).toBeVisible({timeout: 5000})
			// Wait for the specific autocomplete option to be visible (ensures context is updated)
			const workToDoButton = autocompletePopup.getByRole('button', {name: 'Work To Do'})
			await expect(workToDoButton).toBeVisible({timeout: 2000})
			await workToDoButton.click()

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

	test.describe('Edit Saved Filter with Multi-Value Autocomplete (Issue #2010 Regression)', () => {
		test('should preserve filter text after editing and adding trailing space', async ({authenticatedPage: page}) => {
			// This test covers the specific bug from Issue #2010:
			// Creating a filter with 'project in "Work To Do", Inbox', then editing
			// and adding a trailing space should not corrupt the filter or cause errors

			await page.goto('/filters/new')

			// Wait for projects to be loaded
			await expect(page.locator('.menu-list a').filter({hasText: 'Work To Do'})).toBeVisible({timeout: 10000})

			// Step 1: Create a filter with multi-value project using 'in' operator
			await page.locator('input#Title').fill('Work Filter')

			const filterInput = getFilterInput(page)
			await filterInput.click()
			await page.keyboard.press('ControlOrMeta+a')
			await page.keyboard.press('Backspace')

			// Type 'project in Work' to trigger autocomplete
			await filterInput.pressSequentially('project in Work', {delay: 50})

			// Wait for autocomplete popup with the specific option we want to click
			const autocompletePopup = page.locator('#filter-autocomplete-popup')
			await expect(autocompletePopup).toBeVisible({timeout: 5000})
			// Wait for the specific autocomplete option to be visible (ensures context is updated)
			const workToDoButton = autocompletePopup.getByRole('button', {name: 'Work To Do'})
			await expect(workToDoButton).toBeVisible({timeout: 2000})
			await workToDoButton.click()

			// Wait for autocomplete to close and text to stabilize
			await expect(autocompletePopup).not.toBeVisible({timeout: 2000})
			await expect(filterInput).toContainText('project in "Work To Do"')

			// Continue typing the second value: ', Inbox'
			await filterInput.click()
			await page.keyboard.press('End')
			await filterInput.pressSequentially(', Inbox', {delay: 50})

			// Verify the filter text shows the multi-value 'in' clause
			// "Work To Do" should be quoted since it has spaces
			await expect(filterInput).toContainText('project in "Work To Do", Inbox')

			// Step 2: Save the filter and verify no error
			await page.locator('button.is-primary.is-fullwidth').click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()

			// Wait for navigation to the saved filter view
			await expect(page).toHaveURL(/\/projects\/-\d+/, {timeout: 5000})

			// Step 3: Open the filter settings menu from sidebar and click Edit
			// Find the Work Filter link in the sidebar and click its settings menu
			const filterLink = page.locator('.menu-list').getByRole('link', {name: 'Work Filter', exact: true})
			await expect(filterLink).toBeVisible()

			// Hover over the filter to show the settings menu button
			const filterItem = filterLink.locator('..')
			await filterItem.hover()

			// Click the settings menu button
			const settingsButton = filterItem.getByRole('button', {name: 'Open project settings menu'})
			await settingsButton.click()

			// Click "Edit" link in the dropdown menu
			await page.getByRole('link', {name: 'Edit', exact: true}).click()

			// Wait for the edit modal/form to be loaded
			await expect(page.locator('input#Title')).toHaveValue('Work Filter', {timeout: 5000})

			// Find the filter input inside the Filters component
			const editFilterInput = page.locator('.filters .filter-input .ProseMirror')
			await expect(editFilterInput).toBeVisible()

			// Verify the filter text is correctly loaded
			await expect(editFilterInput).toContainText('project in "Work To Do", Inbox')

			// Step 4: Add a trailing space (the bug trigger from #2010)
			await editFilterInput.click()
			// Move to end of text
			await page.keyboard.press('End')
			// Type a space
			await page.keyboard.type(' ')

			// Step 5: Save again and wait for the modal to close (indicates save complete)
			const saveButton = page.locator('.card-footer .button.is-primary')
			await saveButton.click()
			// Wait for the edit card/modal to close after save
			await expect(saveButton).not.toBeVisible({timeout: 5000})

			// Step 6: Assert no error occurred
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()

			// Step 7: Reload and re-open edit to verify the filter text is still intact
			await page.reload()
			await expect(page.locator('.menu-list').getByRole('link', {name: 'Work Filter', exact: true})).toBeVisible({timeout: 5000})

			// Re-open the edit modal
			const filterLinkAfterReload = page.locator('.menu-list').getByRole('link', {name: 'Work Filter', exact: true})
			const filterItemAfterReload = filterLinkAfterReload.locator('..')
			await filterItemAfterReload.hover()
			await filterItemAfterReload.getByRole('button', {name: 'Open project settings menu'}).click()
			await page.getByRole('link', {name: 'Edit', exact: true}).click()

			// Verify the filter title and content are intact
			await expect(page.locator('input#Title')).toHaveValue('Work Filter', {timeout: 5000})

			const reloadedFilterInput = page.locator('.filters .filter-input .ProseMirror')
			// The trailing space may be trimmed, but the core filter should be preserved
			await expect(reloadedFilterInput).toContainText('project in "Work To Do", Inbox')
		})
	})
})
