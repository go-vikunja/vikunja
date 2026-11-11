import {test, expect} from '../../support/fixtures'
import {setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

// Issue #4006: the access token lives 10 minutes, so a browser clock more than that
// off from the server made every fresh token look expired.
const SKEW_MS = 20 * 60 * 1000

test.describe('Login with a skewed browser clock', () => {
	for (const [label, skewMs] of [
		['ahead', SKEW_MS],
		['behind', -SKEW_MS],
	] as const) {
		test(`stays logged in with the clock 20 minutes ${label}`, async ({page, currentUser}) => {
			const {username} = currentUser
			await setupApiUrl(page)
			await page.clock.install({time: new Date(Date.now() + skewMs)})

			await page.goto('/login')
			await page.locator('input[id=username]').fill(username)
			await page.locator('input[id=password]').fill(TEST_PASSWORD)
			await page.locator('.button').filter({hasText: 'Login'}).click()
			await expect(page).toHaveURL('/')
			await expect(page.locator('main h1')).toContainText(username)

			const userRequest = page.waitForRequest(request =>
				request.method() === 'GET' && /\/api\/v[12]\/user$/.test(new URL(request.url()).pathname),
			)
			await page.reload()

			expect((await userRequest).headers()['authorization']).toMatch(/^Bearer .+/)
			await expect(page).toHaveURL('/')
			await expect(page.locator('main h1')).toContainText(username)
		})
	}
})
