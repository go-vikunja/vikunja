import {test, expect} from '../../../support/fixtures'
import {SessionFactory} from '../../../factories/session'
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
})
