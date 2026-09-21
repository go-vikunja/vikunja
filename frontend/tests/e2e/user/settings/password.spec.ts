import {TEST_PASSWORD} from '../../../support/constants'
import {test, expect} from '../../../support/fixtures'

test('changes the password used for subsequent logins', async ({authenticatedPage: page, apiContext, currentUser}) => {
	await page.goto('/user/settings/password-update')
	await page.locator('#password').fill('updatedPassword123')
	await page.locator('#currentPassword').fill(TEST_PASSWORD)
	await page.getByRole('button', {name: 'Save', exact: true}).click()
	await expect(page.locator('.global-notification')).toContainText('Success')
	const login = await apiContext.post('login', {data: {username: currentUser.username, password: 'updatedPassword123'}})
	expect(login.ok()).toBe(true)
	const rejected = await apiContext.post('login', {data: {username: currentUser.username, password: TEST_PASSWORD}})
	expect(rejected.ok()).toBe(false)
})
