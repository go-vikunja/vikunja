import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {SavedFilterFactory} from '../../factories/saved_filter'

/**
 * Tests for filter autocomplete functionality, specifically for:
 * - Project names with spaces (Issue #2010)
 * - Autocomplete selection replacing correct text
 * - Multi-value operators (in, ?=)
 */

async function createProjectWithViews(id: number, title: string) {
	await ProjectFactory.create(1, {
		id,
		title,
	})
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
	test.beforeEach(async () => {
		await ProjectFactory.truncate()
		await TaskFactory.truncate()
		await ProjectViewFactory.truncate()
		await SavedFilterFactory.truncate()

		// Create projects - one with spaces in name (the bug case)
		await createProjectWithViews(1, 'Inbox')
		await createProjectWithViews(2, 'Work To Do')
		await createProjectWithViews(3, 'Personal Tasks')

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

	test.describe('Saved Filter Creation', () => {
		test('should create filter with single-word project name', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			// Fill in filter name
			await page.locator('input#Title').fill('Inbox Filter')

			// Click on filter input and type filter
			const filterInput = getFilterInput(page)
			await filterInput.click()
			await filterInput.press('Control+a')
			await filterInput.pressSequentially('project in Inbox', {delay: 30})

			// Wait for autocomplete to appear
			await expect(page.locator('#filter-autocomplete-popup')).toBeVisible({timeout: 5000})

			// Click the autocomplete suggestion
			await page.locator('#filter-autocomplete-popup button').filter({hasText: 'Inbox'}).click()

			// Verify the filter text is correct (not corrupted)
			await expect(filterInput).toContainText('project in Inbox')

			// Save the filter
			await page.locator('button').filter({hasText: /erstellen|create/i}).click()

			// Verify filter was saved and shows correct results
			await expect(page.locator('.tasks')).toContainText('Inbox Task')
		})

		test('should create filter with multi-word project name containing spaces', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			// Fill in filter name
			await page.locator('input#Title').fill('Work Filter')

			// Click on filter input and type filter
			const filterInput = getFilterInput(page)
			await filterInput.click()
			await filterInput.press('Control+a')
			await filterInput.pressSequentially('project in Work', {delay: 30})

			// Wait for autocomplete to appear
			await expect(page.locator('#filter-autocomplete-popup')).toBeVisible({timeout: 5000})

			// Click the "Work To Do" suggestion (multi-word project name)
			await page.locator('#filter-autocomplete-popup button').filter({hasText: 'Work To Do'}).click()

			// CRITICAL: Verify the filter text is NOT corrupted
			// Before fix: "project in Work To Do, Do" (corrupted)
			// After fix: "project in Work To Do" (correct)
			await expect(filterInput).toContainText('project in Work To Do')
			await expect(filterInput).not.toContainText('Work To Do, Do')
			await expect(filterInput).not.toContainText('Work To Do Do')

			// Save the filter
			await page.locator('button').filter({hasText: /erstellen|create/i}).click()

			// Verify no error message appears
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()

			// Verify filter was saved and shows correct results
			await expect(page.locator('.tasks')).toContainText('Work Task 1')
			await expect(page.locator('.tasks')).toContainText('Work Task 2')
		})

		test('should handle filter with done condition and multi-word project', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			await page.locator('input#Title').fill('Complex Filter')

			const filterInput = getFilterInput(page)
			await filterInput.click()
			await filterInput.press('Control+a')
			await filterInput.pressSequentially('done = false && project in Work', {delay: 30})

			// Wait for autocomplete
			await expect(page.locator('#filter-autocomplete-popup')).toBeVisible({timeout: 5000})

			// Select "Work To Do"
			await page.locator('#filter-autocomplete-popup button').filter({hasText: 'Work To Do'}).click()

			// Verify correct filter text
			await expect(filterInput).toContainText('done = false && project in Work To Do')

			// Save and verify
			await page.locator('button').filter({hasText: /erstellen|create/i}).click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
		})
	})

	test.describe('Saved Filter Editing', () => {
		test('should edit existing filter with multi-word project without corruption', async ({authenticatedPage: page}) => {
			// Create a saved filter via factory
			await SavedFilterFactory.create(1, {
				id: 1,
				title: 'Test Edit Filter',
				filters: JSON.stringify({
					filter: 'project_id in 2',
					filter_include_nulls: true,
					s: '',
				}),
			})

			// Navigate to edit the filter
			await page.goto('/projects/-1/settings/edit')

			// Verify the filter shows correctly (project_id 2 = "Work To Do")
			const filterInput = getFilterInput(page)
			await expect(filterInput).toContainText('project in Work To Do')

			// Click at the end of the filter to trigger autocomplete
			await filterInput.click()
			await filterInput.press('End')

			// If autocomplete appears, select the project
			const autocomplete = page.locator('#filter-autocomplete-popup')
			if (await autocomplete.isVisible({timeout: 2000}).catch(() => false)) {
				await page.locator('#filter-autocomplete-popup button').filter({hasText: 'Work To Do'}).click()
			}

			// Verify filter is not corrupted
			await expect(filterInput).toContainText('project in Work To Do')
			await expect(filterInput).not.toContainText('Work To Do, Do')

			// Save the filter
			await page.locator('button').filter({hasText: /speichern|save/i}).click()

			// Verify no error
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
			await expect(page.locator('.notification.is-success')).toBeVisible()
		})
	})

	test.describe('Multi-value Operators', () => {
		test('should handle multiple projects with in operator', async ({authenticatedPage: page}) => {
			await page.goto('/filters/new')

			await page.locator('input#Title').fill('Multi Project Filter')

			const filterInput = getFilterInput(page)
			await filterInput.click()
			await filterInput.press('Control+a')
			await filterInput.pressSequentially('project in Inbox, Work', {delay: 30})

			// Wait for autocomplete
			await expect(page.locator('#filter-autocomplete-popup')).toBeVisible({timeout: 5000})

			// Select "Work To Do"
			await page.locator('#filter-autocomplete-popup button').filter({hasText: 'Work To Do'}).click()

			// Verify the multi-value filter is correct
			// Should replace "Work" with "Work To Do", keeping "Inbox, "
			await expect(filterInput).toContainText('project in Inbox, Work To Do')

			// Save and verify no error
			await page.locator('button').filter({hasText: /erstellen|create/i}).click()
			await expect(page.locator('.notification.is-danger')).not.toBeVisible()
		})
	})
})
