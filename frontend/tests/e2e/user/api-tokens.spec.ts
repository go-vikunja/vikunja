import {test, expect} from '../../support/fixtures'

test.describe('API Tokens', () => {
	test('Pre-populates title from query parameter', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/api-tokens?title=My%20Test%20Token')
		await page.waitForLoadState('networkidle')

		// Form should be visible automatically
		const titleInput = page.locator('#apiTokenTitle')
		await expect(titleInput).toBeVisible({timeout: 5000})

		// Title should be pre-populated
		await expect(titleInput).toHaveValue('My Test Token')
	})

	test('Pre-selects scopes from query parameter', async ({authenticatedPage: page}) => {
		// Use actual scope names: tasks:create
		await page.goto('/user/settings/api-tokens?scopes=tasks:create')
		await page.waitForLoadState('networkidle')

		// Form should be visible automatically when scopes are provided
		const permissionsLabel = page.locator('label.label:has-text("Permissions")')
		await expect(permissionsLabel).toBeVisible({timeout: 5000})

		// The title input should be visible (form is shown)
		const titleInput = page.locator('#apiTokenTitle')
		await expect(titleInput).toBeVisible()

		// Find the div containing the "tasks" group (has the bold group header checkbox)
		const tasksGroupDiv = page.locator('.mbe-2').filter({
			has: page.locator('.fancy-checkbox.has-text-weight-bold:has-text("tasks")'),
		})
		await expect(tasksGroupDiv).toBeVisible()

		// Within that group, find the specific "create" permission checkbox and verify it's checked
		const createCheckbox = tasksGroupDiv.locator('.fancy-checkbox.mis-4:has-text("create") input[type="checkbox"]')
		await expect(createCheckbox).toBeChecked()
	})

	test('Pre-populates both title and scopes from query parameters', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/api-tokens?title=Integration%20Token&scopes=labels:create')
		await page.waitForLoadState('networkidle')

		// Form should be visible automatically
		const titleInput = page.locator('#apiTokenTitle')
		await expect(titleInput).toBeVisible({timeout: 5000})
		await expect(titleInput).toHaveValue('Integration Token')

		// Permissions section should be visible
		const permissionsLabel = page.locator('label.label:has-text("Permissions")')
		await expect(permissionsLabel).toBeVisible()
	})

	test('Shows create form without query parameters', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/api-tokens')
		await page.waitForLoadState('networkidle')

		// Form should NOT be visible initially
		const titleInput = page.locator('#apiTokenTitle')
		await expect(titleInput).not.toBeVisible({timeout: 2000})

		// Click the create button to show the form
		const createButton = page.locator('button:has-text("Create a token")')
		await expect(createButton).toBeVisible()
		await createButton.click()

		// Now the form should be visible
		await expect(titleInput).toBeVisible()
	})
})
