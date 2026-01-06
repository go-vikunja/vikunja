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
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// The saved filter should appear in the sidebar (as a pseudo-project with negative ID)
		// Saved filters section shows filters that aren't favorites
		const filterItem = page.locator('.menu-container').getByRole('listitem').filter({hasText: 'My Test Filter'})
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

		// The filter should now appear in the Favorites section
		await expect(page.locator('.menu-container')).toContainText('Favorites', {timeout: 10000})

		// Verify the star is now filled (is-favorite class)
		await filterItem.hover()
		await expect(filterItem.locator('.favorite.is-favorite')).toBeVisible()
	})

	test('Can remove a saved filter from favorites', async ({authenticatedPage: page}) => {
		// Create a project (required for saved filters to work)
		await ProjectFactory.create(1)

		// Create a saved filter that is already a favorite
		await SavedFilterFactory.create(1, {
			title: 'Favorited Filter',
			is_favorite: true,
		})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// The saved filter should appear in the Favorites section
		await expect(page.locator('.menu-container')).toContainText('Favorites', {timeout: 10000})

		const filterItem = page.locator('.menu-container').getByRole('listitem').filter({hasText: 'Favorited Filter'})
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

		// The filter should no longer have the is-favorite class
		await filterItem.hover()
		await expect(filterItem.locator('.favorite:not(.is-favorite)')).toBeVisible()
	})

	test('Saved filter favorite status persists after page reload', async ({authenticatedPage: page}) => {
		// Create a project
		await ProjectFactory.create(1)

		// Create a non-favorite saved filter
		await SavedFilterFactory.create(1, {
			title: 'Persistent Filter',
			is_favorite: false,
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

		// Reload the page
		await page.reload()
		await page.waitForLoadState('networkidle')

		// The filter should still be in favorites after reload
		await expect(page.locator('.menu-container')).toContainText('Favorites', {timeout: 10000})
		const reloadedFilterItem = page.locator('.menu-container').getByRole('listitem').filter({hasText: 'Persistent Filter'})
		await reloadedFilterItem.hover()
		await expect(reloadedFilterItem.locator('.favorite.is-favorite')).toBeVisible()
	})
})
