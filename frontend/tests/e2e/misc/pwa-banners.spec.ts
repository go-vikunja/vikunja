import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

// The add to home screen banner is hidden from tablet width up
const iPhone8 = {
	width: 375,
	height: 667,
}

test.describe('PWA banners', () => {
	test.use({viewport: iPhone8})

	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
		await page.goto('/login')
		await expect(page.locator('input[id=username]')).toBeVisible()
		// Service workers are blocked in e2e, so fake the event registerServiceWorker.ts dispatches
		await page.evaluate(() => document.dispatchEvent(new CustomEvent('swUpdated', {detail: {waiting: null}})))
	})

	test('Are hidden on the login page', async ({page}) => {
		await expect(page.locator('.update-notification')).toBeHidden()
		await expect(page.locator('.add-to-home-screen')).toBeHidden()
	})

	test('Show after logging in, including an update found on the login page', async ({page}) => {
		const [user] = await UserFactory.create(1)
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()
		await expect(page).toHaveURL('/')

		await expect(page.locator('.update-notification')).toBeVisible()
		await expect(page.locator('.add-to-home-screen')).toBeVisible()
	})
})
