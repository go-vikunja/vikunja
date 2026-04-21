import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'

test.describe('CalDAV', () => {
	test('generates a token that authenticates against the caldav endpoint', async ({
		authenticatedPage: page, currentUser, apiContext,
	}) => {
		await gotoUserSettings(page, 'caldav')

		const created = page.waitForResponse(r =>
			r.url().includes('/user/settings/token/caldav') && r.request().method() === 'PUT',
		)
		await page.getByRole('button', {name: 'Create a CalDAV token'}).click()
		await created

		// Banner renders the one-time token string; capture it.
		const banner = page.locator('.message').filter({hasText: 'Here is your new token'})
		await expect(banner).toBeVisible()
		const tokenText = await banner.innerText()
		const tokenMatch = tokenText.match(/[A-Za-z0-9]{40,}/)
		expect(tokenMatch).not.toBeNull()
		const token = tokenMatch![0]

		// Hit the caldav principal endpoint with basic auth.
		const basic = Buffer.from(`${currentUser.username}:${token}`).toString('base64')
		const resp = await apiContext.fetch(`/dav/principals/${currentUser.username}/`, {
			method: 'PROPFIND',
			headers: {Authorization: `Basic ${basic}`, Depth: '0'},
		})
		expect(resp.status()).toBeLessThan(300)
	})
})
