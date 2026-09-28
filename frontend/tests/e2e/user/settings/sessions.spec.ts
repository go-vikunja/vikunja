import {test, expect} from '../../../support/fixtures'
import {SessionFactory, hashSessionToken} from '../../../factories/session'
import {gotoUserSettings} from '../../../support/userSettings'

test.describe('Sessions', () => {
	test('lists the current session and other sessions', async ({
		authenticatedPage: page, currentUser,
	}) => {
		// The auth fixture already created one session row (the login).
		// Seed one additional session with truncate=false so we don't wipe it.
		await SessionFactory.create(1, {
			user_id: currentUser.id,
			device_info: 'Firefox on Linux',
			ip_address: '192.0.2.5',
		}, false)

		await gotoUserSettings(page, 'sessions')
		const rows = page.locator('table.table tbody tr')
		await expect(rows).toHaveCount(2)
		await expect(page.locator('.tag.is-primary')).toContainText('Current')
		await expect(page.locator('tr', {hasText: 'Firefox on Linux'})).toContainText('192.0.2.5')
	})

	test('revoking a session breaks its refresh token', async ({
		authenticatedPage: page, currentUser, apiContext,
	}) => {
		const rawToken = 'fixed-refresh-token-for-test-12345678901234567890'
		await SessionFactory.create(1, {
			user_id: currentUser.id,
			token_hash: hashSessionToken(rawToken),
			ip_address: '192.0.2.5',
			device_info: 'Firefox on Linux',
		}, false)

		await gotoUserSettings(page, 'sessions')
		await page.locator('tr', {hasText: /192\.0\.2\.5/})
			.getByRole('button', {name: 'Delete'}).click()
		const deleted = page.waitForResponse(r =>
			/\/user\/sessions\/[^/]+/.test(r.url()) && r.request().method() === 'DELETE',
		)
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await deleted
		await expect(page.locator('table.table tbody tr')).toHaveCount(1)

		// After revoke, the refresh request must fail. Refresh tokens live in the
		// vikunja_refresh_token cookie, not as a Bearer credential.
		const after = await apiContext.post('user/token/refresh', {
			headers: {Cookie: `vikunja_refresh_token=${rawToken}`},
		})
		expect(after.status()).toBe(401)
		await page.reload()
		await expect(page.locator('table.table tbody tr')).toHaveCount(1)
	})

	test('pages the session list and steps back after revoking the last row of a page', async ({
		authenticatedPage: page, currentUser,
	}) => {
		await SessionFactory.create(25, {
			user_id: currentUser.id,
			ip_address: (i: number) => `192.0.2.${i}`,
			last_active: (i: number) => new Date(Date.now() - i * 60_000).toISOString(),
		}, false)

		await gotoUserSettings(page, 'sessions')
		const rows = page.locator('table.table tbody tr')
		await expect(rows).toHaveCount(25)
		await expect(page.locator('.tag.is-primary')).toContainText('Current')

		await page.getByRole('link', {name: 'Goto page 2'}).click()
		await expect(page).toHaveURL(/[?&]page=2/)
		await expect(rows).toHaveCount(1)
		await expect(rows).toContainText('192.0.2.25')

		await rows.getByRole('button', {name: 'Delete'}).click()
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await expect(page).not.toHaveURL(/[?&]page=2/)
		await expect(rows).toHaveCount(25)
		await expect(page.getByRole('navigation', {name: 'pagination'})).toHaveCount(0)

		await page.reload()
		await expect(rows).toHaveCount(25)
		await expect(page.locator('tr', {hasText: '192.0.2.25'})).toHaveCount(0)
	})

	test('current session cannot be deleted from the UI', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'sessions')
		const currentRow = page.locator('tr', {has: page.locator('.tag.is-primary')})
		await expect(currentRow.getByRole('button', {name: 'Delete'})).toHaveCount(0)
	})
})
