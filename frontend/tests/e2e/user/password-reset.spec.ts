import {test, expect} from '../../support/fixtures'
import {UserFactory, type UserAttributes} from '../../factories/user'
import {TokenFactory, type TokenAttributes} from '../../factories/token'

test.describe('Password Reset', () => {
	let user: UserAttributes

	test.beforeEach(async ({page, apiContext}) => {
		await UserFactory.truncate()
		await TokenFactory.truncate()
		const users = await UserFactory.create(1)
		user = users[0] as UserAttributes
	})

	test('Should allow a user to reset their password with a valid token', async ({page, apiContext}) => {
		const tokenArray = await TokenFactory.create(1, {user_id: user.id as number, kind: 1})
		const token: TokenAttributes = tokenArray[0] as TokenAttributes

		await page.goto(`/?userPasswordReset=${token.token}`)
		await expect(page).toHaveURL(`/password-reset?userPasswordReset=${token.token}`)

		const newPassword = 'newSecurePassword123'
		await page.locator('input[id=password]').fill(newPassword)
		await page.locator('button').filter({hasText: 'Reset your password'}).click()

		await expect(page.locator('.message.success')).toContainText('The password was updated successfully.')
		await page.locator('.button').filter({hasText: 'Login'}).click()
		await expect(page).toHaveURL('/login')

		// Try to login with the new password
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(newPassword)
		await page.locator('.button').filter({hasText: 'Login'}).click()
		await expect(page).toHaveURL('/')
	})

	test('Should show an error for an invalid token', async ({page, apiContext}) => {
		await page.goto('/?userPasswordReset=invalidtoken123')
		await expect(page).toHaveURL('/password-reset?userPasswordReset=invalidtoken123')

		// Attempt to reset password
		const newPassword = 'newSecurePassword123'
		await page.locator('input[id=password]').fill(newPassword)
		await page.locator('button').filter({hasText: 'Reset your password'}).click()

		await expect(page.locator('.message')).toContainText('Invalid token')
	})

	test('Should redirect to login if no token is present in query param when visiting /password-reset directly', async ({page, apiContext}) => {
		await page.goto('/password-reset')
		// Wait for redirect to login page
		await expect(page).toHaveURL('/login')
	})

	test('Should redirect to login if userPasswordReset token is not present in query param when visiting root', async ({page, apiContext}) => {
		await page.goto('/')
		await expect(page).toHaveURL('/login')
	})
})
