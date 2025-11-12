import {test, expect} from '../../support/fixtures'

test.describe('The Menu', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await page.goto('/')
	})

	test('Is visible by default on desktop', async ({authenticatedPage: page}) => {
		await expect(page.locator('.menu-container')).toHaveClass(/is-active/)
	})

	test('Can be hidden on desktop', async ({authenticatedPage: page}) => {
		await page.locator('button.menu-show-button:visible').click()
		await expect(page.locator('.menu-container')).not.toHaveClass(/is-active/)
	})

	test('Is hidden by default on mobile', async ({authenticatedPage: page}) => {
		await page.setViewportSize({width: 375, height: 667}) // iphone-8
		await expect(page.locator('.menu-container')).not.toHaveClass(/is-active/)
	})

	test('Is can be shown on mobile', async ({authenticatedPage: page}) => {
		await page.setViewportSize({width: 375, height: 667}) // iphone-8
		await page.locator('button.menu-show-button:visible').click()
		await expect(page.locator('.menu-container')).toHaveClass(/is-active/)
	})
})
