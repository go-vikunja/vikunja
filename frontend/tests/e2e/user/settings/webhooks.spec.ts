import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'

test('creates and deletes an account webhook with persisted state', async ({
	authenticatedPage: page,
	apiContext,
	userToken,
}) => {
	await gotoUserSettings(page, 'webhooks')
	await page.locator('#targetUrl').fill('https://example.com/account-hook')
	await page.locator('.available-events-check').first().locator('.base-checkbox__label').click()
	await page.getByRole('button', {name: /create webhook/i}).click()
	const row = page.locator('tbody tr', {hasText: 'example.com/account-hook'})
	await expect(row).toBeVisible()
	await page.reload()
	await expect(row).toBeVisible()
	const headers = {Authorization: `Bearer ${userToken}`}
	const stored = await apiContext.get('/api/v2/user/settings/webhooks', {headers})
	expect(stored.ok()).toBe(true)
	const items = (await stored.json()).items
	expect(items).toEqual([expect.objectContaining({target_url: 'https://example.com/account-hook'})])
	await row.locator('.button.is-danger').click()
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(row).toHaveCount(0)
	await page.reload()
	await expect(row).toHaveCount(0)
	const deleted = await apiContext.get('/api/v2/user/settings/webhooks', {headers})
	expect((await deleted.json()).items ?? []).toEqual([])
})
