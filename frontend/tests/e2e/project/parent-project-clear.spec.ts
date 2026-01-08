import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {createDefaultViews} from './prepareProjects'

test.describe('Parent Project Clear', () => {
	test('Should clear the parent project field and persist the change', async ({authenticatedPage: page}) => {
		// Create a parent project
		const parentProjects = await ProjectFactory.create(1, {
			id: 100,
			title: 'Parent Project',
		})
		await createDefaultViews(parentProjects[0].id, 100)

		// Create a child project with the parent
		const childProjects = await ProjectFactory.create(1, {
			id: 101,
			title: 'Child Project',
			parent_project_id: parentProjects[0].id,
		}, false)
		const childViews = await createDefaultViews(childProjects[0].id, 104, false)

		// Navigate to the child project first
		await page.goto(`/projects/${childProjects[0].id}/${childViews[0].id}`)
		await page.waitForLoadState('networkidle')
		await expect(page.locator('.project-title')).toContainText('Child Project')

		// Open project settings dropdown and click Edit
		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^edit$/i}).click()
		await page.waitForLoadState('networkidle')

		// Verify the parent project is shown in the modal
		const parentProjectInput = page.locator('.multiselect input')
		await expect(parentProjectInput).toHaveValue('Parent Project')

		// Click the clear button (X) on the parent project field
		await page.locator('.multiselect .removal-button').click()

		// Verify the field is cleared (should show empty/placeholder)
		await expect(parentProjectInput).toHaveValue('')

		// Save the project
		await page.locator('footer.card-footer .button').filter({hasText: /^Save$/}).click()
		await expect(page.locator('.global-notification')).toContainText('Success')

		// Verify the project is no longer nested in the sidebar
		// Child Project should now be a top-level item, not inside Parent Project's subtree
		const sidebar = page.locator('.menu-container .menu-list')
		// The Child Project link should be a direct child of the sidebar, not nested under Parent Project
		await expect(sidebar.getByRole('link', {name: 'Child Project'})).toBeVisible()
		// Verify Child Project is NOT inside the Parent Project's nested list
		const parentProjectItem = sidebar.getByRole('listitem').filter({has: page.getByRole('link', {name: 'Parent Project'})})
		await expect(parentProjectItem.getByRole('link', {name: 'Child Project'})).not.toBeVisible()

		// Open edit again to verify parent is still cleared
		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^edit$/i}).click()
		await page.waitForLoadState('networkidle')

		// Verify the parent project field is still empty
		await expect(parentProjectInput).toHaveValue('')
	})

	test('Should not jump back after selecting and deleting the parent project text', async ({authenticatedPage: page}) => {
		// Create a parent project
		const parentProjects = await ProjectFactory.create(1, {
			id: 200,
			title: 'Test Parent',
		})
		await createDefaultViews(parentProjects[0].id, 200)

		// Create a child project with the parent
		const childProjects = await ProjectFactory.create(1, {
			id: 201,
			title: 'Test Child',
			parent_project_id: parentProjects[0].id,
		}, false)
		const childViews = await createDefaultViews(childProjects[0].id, 204, false)

		// Navigate to the child project first
		await page.goto(`/projects/${childProjects[0].id}/${childViews[0].id}`)
		await page.waitForLoadState('networkidle')
		await expect(page.locator('.project-title')).toContainText('Test Child')

		// Open project settings dropdown and click Edit
		await page.locator('.project-title-dropdown .project-title-button').click()
		await page.getByRole('link', {name: /^edit$/i}).click()
		await page.waitForLoadState('networkidle')

		const parentProjectInput = page.locator('.multiselect input')

		// Verify the parent project is shown
		await expect(parentProjectInput).toHaveValue('Test Parent')

		// Select all text and delete it (simulating user manually clearing the field)
		await parentProjectInput.click()
		await parentProjectInput.selectText()
		await page.keyboard.press('Backspace')

		// Wait a moment to ensure the value doesn't jump back
		await page.waitForTimeout(500)

		// Verify the field stays empty (this was the bug - it would jump back)
		await expect(parentProjectInput).toHaveValue('')
	})
})
