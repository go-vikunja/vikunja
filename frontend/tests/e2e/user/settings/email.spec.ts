import {test, expect} from '../../../support/fixtures'
import {TEST_PASSWORD} from '../../../support/constants'

test('stores an email update when mail confirmation is disabled', async ({authenticatedPage: page, apiContext, userToken}) => {
	await page.goto('/user/settings/email-update')
	await page.locator('#newEmail').fill('changed@example.test')
	await page.locator('#currentPasswordEmail').fill(TEST_PASSWORD)
	await page.getByRole('button', {name: 'Save', exact: true}).click()
	await expect(page.locator('.global-notification')).toContainText('Success')
	await page.reload()
	const response = await apiContext.get('user', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await response.json()).email).toBe('changed@example.test')
})
