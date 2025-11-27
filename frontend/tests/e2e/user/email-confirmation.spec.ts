import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TokenFactory} from '../../factories/token'
import {TEST_PASSWORD, TEST_PASSWORD_HASH} from '../../support/constants'

test.describe('Email Confirmation', () => {
	let user
	let confirmationToken

	test.beforeEach(async ({page, apiContext}) => {
		await UserFactory.truncate()
		await TokenFactory.truncate()

		// Create a user with status = 1 (StatusEmailConfirmationRequired)
		const users = await UserFactory.create(1, {
			username: 'unconfirmeduser',
			email: 'unconfirmed@example.com',
			password: TEST_PASSWORD_HASH,
			status: 1, // StatusEmailConfirmationRequired
		})
		user = users[0]

		// Create an email confirmation token for this user
		// kind: 2 = TokenEmailConfirm
		confirmationToken = 'test-email-confirm-token-12345678901234567890123456789012'
		await TokenFactory.create(1, {
			user_id: user.id,
			kind: 2,
			token: confirmationToken,
		})
	})

	test('Should fail login before email is confirmed', async ({page, apiContext}) => {
		await page.goto('/login')
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		await expect(page.locator('div.message.danger')).toContainText('Email address of the user not confirmed')
	})

	test('Should confirm email and allow login', async ({page, apiContext}) => {
		// Setup response promise for the confirmation API call
		const confirmEmailPromise = page.waitForResponse(response =>
			response.url().includes('/user/confirm') && response.request().method() === 'POST',
		)

		// Manually set the token in localStorage before visiting the page
		// This simulates what happens when the user clicks the email link
		await page.goto('/login')
		await page.evaluate((token) => {
			localStorage.setItem('emailConfirmToken', token)
		}, confirmationToken)
		await page.reload()

		// Wait for the confirmation API call to complete
		const confirmResponse = await confirmEmailPromise
		expect(confirmResponse.status()).toBe(200)

		// Should show success message
		await expect(page.locator('.message.success')).toBeVisible({timeout: 10000})
		await expect(page.locator('.message.success')).toContainText('You successfully confirmed your email')

		// Now login should work
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		// Should successfully log in
		await expect(page).toHaveURL(/\//)
		await expect(page).not.toHaveURL(/\/login/)
		// Check that the username appears in the greeting
		await expect(page.locator('body')).toContainText(user.username)
	})

	test('Should fail with invalid confirmation token', async ({page, apiContext}) => {
		// Setup response promise for the confirmation API call
		const confirmEmailPromise = page.waitForResponse(response =>
			response.url().includes('/user/confirm') && response.request().method() === 'POST',
		)

		// Try to confirm with an invalid token
		const invalidToken = 'invalid-token-that-does-not-exist-in-database'
		await page.goto('/login')
		await page.evaluate((token) => {
			localStorage.setItem('emailConfirmToken', token)
		}, invalidToken)
		await page.reload()

		// Wait for the confirmation API call to fail
		await confirmEmailPromise

		// Should show error message
		await expect(page.locator('.message.danger')).toBeVisible({timeout: 10000})

		// Login should still fail
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		await expect(page.locator('div.message.danger')).toContainText('Email address of the user not confirmed')
	})

	test('Should not allow using the same token twice', async ({page, apiContext}) => {
		// First confirmation - should work
		let confirmEmailPromise = page.waitForResponse(response =>
			response.url().includes('/user/confirm') && response.request().method() === 'POST',
		)

		await page.goto('/login')
		await page.evaluate((token) => {
			localStorage.setItem('emailConfirmToken', token)
		}, confirmationToken)
		await page.reload()

		let confirmResponse = await confirmEmailPromise
		expect(confirmResponse.status()).toBe(200)
		await expect(page.locator('.message.success')).toBeVisible({timeout: 10000})
		await expect(page.locator('.message.success')).toContainText('You successfully confirmed your email')

		// Try to use the same token again - should fail
		confirmEmailPromise = page.waitForResponse(response =>
			response.url().includes('/user/confirm') && response.request().method() === 'POST',
		)

		await page.goto('/login')
		await page.evaluate((token) => {
			localStorage.setItem('emailConfirmToken', token)
		}, confirmationToken)
		await page.reload()

		await confirmEmailPromise
		await expect(page.locator('.message.danger')).toBeVisible({timeout: 10000})
	})

	test('Should confirm email when clicking link from email (via query parameter)', async ({page, apiContext}) => {
		// Setup response promise for the confirmation API call
		const confirmEmailPromise = page.waitForResponse(response =>
			response.url().includes('/user/confirm') && response.request().method() === 'POST',
		)

		// Simulate clicking the email confirmation link with query parameter
		// This is what happens when a user clicks the link in their email
		await page.goto(`/?userEmailConfirm=${confirmationToken}`)

		// Should redirect to login page
		await expect(page).toHaveURL(/\/login/)

		// Wait for the confirmation API call to complete
		const confirmResponse = await confirmEmailPromise
		expect(confirmResponse.status()).toBe(200)

		// Should show success message
		await expect(page.locator('.message.success')).toBeVisible({timeout: 10000})
		await expect(page.locator('.message.success')).toContainText('You successfully confirmed your email')

		// Now login should work
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		// Should successfully log in
		await expect(page).toHaveURL(/\//)
		await expect(page).not.toHaveURL(/\/login/)
		// Check that the username appears in the greeting
		await expect(page.locator('body')).toContainText(user.username)
	})
})
