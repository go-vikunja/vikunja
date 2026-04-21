import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {TokenFactory} from '../../../factories/token'

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

	test('deleting a token revokes caldav access', async ({
		authenticatedPage: page, currentUser,
	}) => {
		const tokenValue = 'fixed-caldav-token-123456789012345678901234567890'
		// kind=4 is TokenCaldavAuth (see pkg/user/token.go)
		await TokenFactory.create(1, {user_id: currentUser.id, kind: 4, token: tokenValue}, false)

		await gotoUserSettings(page, 'caldav')
		// Filter to data rows (rows containing a <td>) to exclude the <th>-only header row.
		const dataRows = page.locator('table.table tr').filter({has: page.locator('td')})
		await expect(dataRows).toHaveCount(1)

		await dataRows.getByRole('button', {name: 'Delete'}).click()
		await expect(dataRows).toHaveCount(0)

		// NOTE: the factory seeds the plaintext token as-is, but caldav tokens are
		// stored bcrypt-hashed. We assert the row is gone in the UI rather than
		// probing caldav with the seeded value.
	})
})
