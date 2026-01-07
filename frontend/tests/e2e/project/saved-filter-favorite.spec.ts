import {test, expect} from '../../support/fixtures'
import {SavedFilterFactory} from '../../factories/saved_filter'
import {ProjectFactory} from '../../factories/project'

test.describe('Saved Filter Favorites', () => {
	test('Can mark a saved filter as favorite', async ({authenticatedPage: page}) => {
		// Create a project (required for saved filters to work)
		await ProjectFactory.create(1)

		// Create a saved filter (not favorite initially)
		await SavedFilterFactory.create(1, {
			title: 'My Test Filter',
			is_favorite: false,
			filters: '{"filter":"done = false","filter_include_nulls":false,"s":""}',
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// The saved filter should appear in the sidebar (as a pseudo-project with negative ID)
		// Saved filters section shows filters that aren't favorites
		const filterItem = page.locator('.list-menu .navigation-item').filter({hasText: 'My Test Filter'})
		await expect(filterItem).toBeVisible({timeout: 10000})

		// Hover to reveal the favorite button
		await filterItem.hover()
		const favoriteButton = filterItem.locator('.favorite')
		await expect(favoriteButton).toBeVisible()

		// Click to mark as favorite
		const favoritePromise = page.waitForResponse(response =>
			response.url().includes('/filters/') && response.request().method() === 'POST',
		)
		await favoriteButton.click()
		await favoritePromise

		await expect(filterItem.locator('.favorite.is-favorite').nth(0)).toBeVisible()
		await expect(filterItem.locator('.favorite.is-favorite').nth(1)).toBeVisible()
	})

	test('Can remove a saved filter from favorites', async ({authenticatedPage: page}) => {
		// Create a project (required for saved filters to work)
		await ProjectFactory.create(1)

		// Create a saved filter that is already a favorite
		await SavedFilterFactory.create(1, {
			title: 'Favorited Filter',
			is_favorite: true,
			filters: '{"filter":"done = false","filter_include_nulls":false,"s":""}',
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// The saved filter appears twice (favorites section + saved filters section)
		// Get the first instance (in favorites section) which should have a filled star
		const filterItem = page.locator('.menu-container').getByRole('listitem').filter({hasText: 'Favorited Filter'}).first()
		await expect(filterItem).toBeVisible()

		// Hover to reveal the favorite button (should be filled star)
		await filterItem.hover()
		const favoriteButton = filterItem.locator('.favorite.is-favorite')
		await expect(favoriteButton).toBeVisible()

		// Click to remove from favorites
		const unfavoritePromise = page.waitForResponse(response =>
			response.url().includes('/filters/') && response.request().method() === 'POST',
		)
		await favoriteButton.click()
		await unfavoritePromise

		// After unfavoriting, the star should be outline (not filled)
		// Wait for UI to update with longer timeout
		await expect(filterItem.locator('.favorite:not(.is-favorite)')).toBeVisible({timeout: 10000})
	})

	test('Saved filter favorite status persists after page reload', async ({authenticatedPage: page}) => {
		// Create a project
		await ProjectFactory.create(1)

		// Create a non-favorite saved filter
		await SavedFilterFactory.create(1, {
			title: 'Persistent Filter',
			is_favorite: false,
			filters: '{"filter":"done = false","filter_include_nulls":false,"s":""}',
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Find and favorite the filter
		const filterItem = page.locator('.menu-container').getByRole('listitem').filter({hasText: 'Persistent Filter'})
		await filterItem.hover()

		const favoritePromise = page.waitForResponse(response =>
			response.url().includes('/filters/') && response.request().method() === 'POST',
		)
		await filterItem.locator('.favorite').click()
		await favoritePromise

		// Wait for UI to update before reloading
		await page.waitForTimeout(500)

		// Reload the page
		await page.reload()
		await page.waitForLoadState('networkidle')

		// The filter should still be favorited after reload (filled star)
		await expect(filterItem.locator('.favorite.is-favorite').nth(0)).toBeVisible()
		await expect(filterItem.locator('.favorite.is-favorite').nth(1)).toBeVisible()
	})
})
