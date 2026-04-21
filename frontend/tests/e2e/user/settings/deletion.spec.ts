import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {TEST_PASSWORD} from '../../../support/constants'
import {TokenFactory} from '../../../factories/token'

test.describe('Account deletion', () => {
	test('blocks deletion request with no password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'deletion')
		await page.locator('.card .button.is-danger').click()
		await expect(page.locator('.help.is-danger')).toContainText(/password/i)
	})

	test('schedules deletion with correct password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'deletion')
		await page.locator('#currentPasswordAccountDelete').fill(TEST_PASSWORD)

		const resp = page.waitForResponse(r => r.url().includes('/user/deletion/request'))
		await page.locator('.card .button.is-danger').click()
		const r = await resp
		expect(r.ok()).toBe(true)
		await expect(page.locator('.global-notification .vue-notification.success')).toBeVisible()
	})

	test('cancels a scheduled deletion', async ({
		authenticatedPage: page, currentUser, userToken, apiContext,
	}) => {
		const deletionToken = 'fixed-account-deletion-token-1234567890123456'
		// kind=3 is TokenAccountDeletion (see pkg/user/token.go)
		await TokenFactory.create(1, {
			user_id: currentUser.id,
			kind: 3,
			token: deletionToken,
		}, false)

		// Confirm the deletion via API — this is the write that sets deletion_scheduled_at.
		const confirm = await apiContext.post('user/deletion/confirm', {
			headers: {Authorization: `Bearer ${userToken}`},
			data: {token: deletionToken},
		})
		expect(confirm.ok()).toBe(true)

		await gotoUserSettings(page, 'deletion')
		// Scheduled-state copy: "We will delete your Vikunja account at ..."
		await expect(page.locator('.card')).toContainText(/we will delete your Vikunja account/i)

		await page.locator('#currentPasswordAccountDelete').fill(TEST_PASSWORD)
		const cancel = page.waitForResponse(r => r.url().includes('/user/deletion/cancel'))
		await page.getByRole('button', {name: /cancel the deletion/i}).click()
		await cancel

		await expect(page.locator('.global-notification .vue-notification.success')).toBeVisible()
		// And the non-scheduled branch (the "Delete account" form) reappears.
		await expect(page.locator('.card .button.is-danger')).toBeVisible()
	})
})
