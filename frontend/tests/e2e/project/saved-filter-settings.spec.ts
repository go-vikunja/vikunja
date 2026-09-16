import {test, expect} from '../../support/fixtures'
import {SavedFilterFactory} from '../../factories/saved_filter'
import {ProjectFactory} from '../../factories/project'
import type {Page} from '@playwright/test'

// Saved filter id 1 shows up as pseudo-project -2.
async function openFilterSettings(page: Page, item: RegExp) {
	await page.goto('/projects/-2')
	await expect(page.locator('.project-title')).toContainText('My Test Filter')
	await page.locator('.project-title-dropdown .project-title-button').click()
	await page.getByRole('link', {name: item}).click()
}

test.describe('Saved Filter Settings', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
		await SavedFilterFactory.create(1, {
			id: 1,
			title: 'My Test Filter',
			filters: '{"filter":"done = false","filter_include_nulls":false,"s":""}',
		})
	})

	test('Can edit the title of a saved filter', async ({authenticatedPage: page}) => {
		await openFilterSettings(page, /^edit$/i)

		const title = page.locator('#Title')
		await expect(title).toHaveValue('My Test Filter')
		await title.fill('Renamed Filter')

		const updatePromise = page.waitForResponse(response =>
			response.url().includes('/api/v2/filters/1') && response.request().method() === 'PUT',
		)
		await page.locator('footer.card-footer .button').filter({hasText: /^Save$/}).click()
		await updatePromise

		await expect(page.locator('.project-title')).toContainText('Renamed Filter')
		const sidebar = page.locator('.list-menu .navigation-item')
		await expect(sidebar.filter({hasText: 'Renamed Filter'})).toBeVisible()
		await expect(sidebar.filter({hasText: 'My Test Filter'})).toHaveCount(0)
	})

	test('Can delete a saved filter', async ({authenticatedPage: page}) => {
		await openFilterSettings(page, /^delete$/i)

		const deletePromise = page.waitForResponse(response =>
			response.url().includes('/api/v2/filters/1') && response.request().method() === 'DELETE',
		)
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await deletePromise

		await expect(page).toHaveURL(/\/projects$/)
		await expect(page.locator('.list-menu .navigation-item').filter({hasText: 'My Test Filter'})).toHaveCount(0)
	})

	test('Opens the filter edit modal when loading its URL directly', async ({authenticatedPage: page}) => {
		await page.goto('/projects/-2/settings/edit')

		await expect(page.getByText('Edit This Saved Filter')).toBeVisible()
		await expect(page.locator('#Title')).toHaveValue('My Test Filter')
		await expect(page.getByText('Edit This Project')).toHaveCount(0)
	})

	test('Opens the filter delete modal when loading its URL directly', async ({authenticatedPage: page}) => {
		await page.goto('/projects/-2/settings/delete')

		await expect(page.getByText('Delete this saved filter', {exact: true})).toBeVisible()
		await expect(page.getByText('Delete this project', {exact: true})).toHaveCount(0)
	})
})
