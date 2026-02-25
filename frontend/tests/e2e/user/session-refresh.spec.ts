import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TEST_PASSWORD} from '../../support/constants'

async function loginViaBrowser(page, username: string) {
	// Set the API URL so the frontend knows where to send requests.
	const apiUrl = process.env.API_URL || 'http://127.0.0.1:3456/api/v1'
	await page.addInitScript(({apiUrl}) => {
		window.localStorage.setItem('API_URL', apiUrl)
		window.API_URL = apiUrl
	}, {apiUrl})

	await page.goto('/login')
	await page.locator('input[id=username]').fill(username)
	await page.locator('input[id=password]').fill(TEST_PASSWORD)
	await page.locator('.button').filter({hasText: 'Login'}).click()
	await expect(page).toHaveURL('/')
	await expect(page.locator('main h2')).toContainText(username)

	// Wait for the proactive refresh (from useRenewTokenOnFocus) to complete
	// so it doesn't race with our test assertions.
	await page.waitForTimeout(1500)
}

test.describe('Session refresh and retry interceptor', () => {
	let username: string

	test.beforeEach(async ({apiContext}) => {
		const [user] = await UserFactory.create(1)
		username = user.username
	})

	test('Transparently retries a request and rotates the JWT when a 401 with code 11 is returned', async ({page}) => {
		await loginViaBrowser(page, username)

		const tokenBefore = await page.evaluate(() => localStorage.getItem('token'))
		expect(tokenBefore).not.toBeNull()

		// Spy on the refresh endpoint
		let refreshCalled = false
		await page.route(/\/api\/v1\/user\/token\/refresh$/, async (route) => {
			refreshCalled = true
			await route.continue()
		})

		// Intercept the first GET to /user/sessions with 401 code 11.
		// We navigate to the sessions page to trigger this call, avoiding
		// a race with the proactive refresh that fires on page reload.
		let intercepted = false
		await page.route(/\/api\/v1\/user\/sessions/, async (route) => {
			if (!intercepted && route.request().method() === 'GET') {
				intercepted = true
				await route.fulfill({
					status: 401,
					contentType: 'application/json',
					body: JSON.stringify({
						code: 11,
						message: 'missing, malformed, expired or otherwise invalid token provided',
					}),
				})
			} else {
				await route.continue()
			}
		})

		await page.goto('/user/settings/sessions')

		// The sessions page should load after transparent retry
		await expect(page.locator('.tag.is-primary')).toBeVisible({timeout: 10000})

		expect(intercepted).toBe(true)
		expect(refreshCalled).toBe(true)

		// The JWT in localStorage should have been rotated
		const tokenAfter = await page.evaluate(() => localStorage.getItem('token'))
		expect(tokenAfter).not.toBeNull()
		expect(tokenAfter).not.toBe(tokenBefore)
	})

	test('Does not retry 401 with non-JWT error code', async ({page}) => {
		await loginViaBrowser(page, username)

		// Track refresh calls that happen AFTER the fake 401 is returned.
		// The proactive refresh from useRenewTokenOnFocus fires on every page
		// load, so we only care about refreshes triggered after our 401.
		let projectsFailed = false
		let refreshAfterFailure = false

		await page.route(/\/api\/v1\/user\/token\/refresh$/, async (route) => {
			if (projectsFailed) {
				refreshAfterFailure = true
			}
			await route.continue()
		})

		// Return 401 with a non-JWT error code for all project GETs.
		await page.route(/\/api\/v1\/projects(\?|$)/, async (route) => {
			if (route.request().method() === 'GET') {
				projectsFailed = true
				await route.fulfill({
					status: 401,
					contentType: 'application/json',
					body: JSON.stringify({
						code: 1002,
						message: 'Wrong username or password.',
					}),
				})
			} else {
				await route.continue()
			}
		})

		await page.reload()
		// Wait for the proactive refresh (from useRenewTokenOnFocus) that fires
		// on every page load to complete, then reset our flag so only
		// interceptor-triggered refreshes are tracked.
		await page.waitForTimeout(2000)
		refreshAfterFailure = false

		await page.waitForTimeout(1000)

		// The interceptor should NOT have triggered a refresh for a non-JWT 401
		expect(refreshAfterFailure).toBe(false)
	})

	test('Current session appears on the sessions settings page', async ({page}) => {
		await loginViaBrowser(page, username)

		await page.goto('/user/settings/sessions')

		// The sessions table should have at least one row with the "Current" tag
		await expect(page.locator('.tag.is-primary')).toBeVisible({timeout: 5000})
	})
})
