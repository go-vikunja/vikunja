import {login, setupApiUrl} from '../../../support/authenticateUser'
import type {APIRequestContext} from '@playwright/test'
import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {linkDexAccount} from '../../../support/linkedAccounts'
import {UserFactory} from '../../../factories/user'
import {UserIdentityFactory} from '../../../factories/user_identity'
import {TEST_PASSWORD} from '../../../support/constants'

async function listIdentities(apiContext: APIRequestContext, token: string) {
	const response = await apiContext.get('/api/v2/user/settings/identities', {
		headers: {Authorization: `Bearer ${token}`},
	})
	expect(response.ok()).toBe(true)
	return (await response.json()).items ?? []
}

test.describe('Linked accounts', () => {
	test('links a Dex account after confirming the password', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		await gotoUserSettings(page, 'linked-accounts')
		await expect(page.getByText('You have not linked any accounts yet.')).toBeVisible()

		await linkDexAccount(page)
		await expect(page.getByText('test@example.com', {exact: true})).toBeVisible()
		expect(await listIdentities(apiContext, userToken)).toHaveLength(0)
		await page.locator('#linkIdentityPassword').fill(TEST_PASSWORD)
		await page.getByRole('button', {name: 'Link account'}).click()

		await expect(page).toHaveURL(/\/user\/settings\/linked-accounts/)
		await page.reload()
		await expect(page.locator('.card table')).toContainText('Dex')
		await expect(page.getByRole('button', {name: 'Link Dex', exact: true})).toHaveCount(0)

		const identities = await listIdentities(apiContext, userToken)
		expect(identities).toHaveLength(1)
		expect(identities[0].provider_key).toBe('dex')
	})

	test('rejects an account linked to another user before asking for a password', async ({
		authenticatedPage: ownerPage,
		apiContext,
		userToken: ownerToken,
		browser,
	}) => {
		await gotoUserSettings(ownerPage, 'linked-accounts')
		await linkDexAccount(ownerPage)
		await ownerPage.locator('#linkIdentityPassword').fill(TEST_PASSWORD)
		await ownerPage.getByRole('button', {name: 'Link account', exact: true}).click()
		await expect(ownerPage).toHaveURL(/\/user\/settings\/linked-accounts/)

		const [otherUser] = await UserFactory.create(1, {id: 2}, false)
		const context = await browser.newContext()
		try {
			const page = await context.newPage()
			const {token} = await login(page, apiContext, otherUser)
			await gotoUserSettings(page, 'linked-accounts')
			await linkDexAccount(page)
			await expect(page.getByText('This account at the identity provider is already linked to another Vikunja account.', {exact: true})).toBeVisible()
			await expect(page.locator('#linkIdentityPassword')).toHaveCount(0)
			await expect(page.getByRole('link', {name: 'Back to linked accounts', exact: true})).toBeVisible()
			expect(await listIdentities(apiContext, token)).toHaveLength(0)
			expect(await listIdentities(apiContext, ownerToken)).toHaveLength(1)
			expect(await page.evaluate(() => [localStorage.getItem('state'), localStorage.getItem('openidLinking')])).toEqual([null, null])
		} finally {
			await context.close()
		}
	})

	test('keeps the confirmation form after a wrong password', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		await gotoUserSettings(page, 'linked-accounts')
		await linkDexAccount(page)
		await expect(page.getByText('test@example.com', {exact: true})).toBeVisible()
		expect(await listIdentities(apiContext, userToken)).toHaveLength(0)

		await page.locator('#linkIdentityPassword').fill('wrong password')
		await page.getByRole('button', {name: 'Link account'}).click()
		await expect(page.getByText('Wrong username or password.')).toBeVisible()
		expect(await listIdentities(apiContext, userToken)).toHaveLength(0)

		// Reload reuses the prepared identity even though the provider code was exchanged.
		await page.reload()
		await expect(page.getByText('test@example.com', {exact: true})).toBeVisible()
		await page.locator('#linkIdentityPassword').fill(TEST_PASSWORD)
		await page.getByRole('button', {name: 'Link account'}).click()
		await expect(page).toHaveURL(/\/user\/settings\/linked-accounts/)
		expect(await listIdentities(apiContext, userToken)).toHaveLength(1)
	})

	test('unlinks an identity of a removed provider', async ({
		authenticatedPage: page,
		apiContext,
		currentUser,
		userToken,
	}) => {
		await UserIdentityFactory.create(1, {user_id: currentUser.id})
		await gotoUserSettings(page, 'linked-accounts')

		const row = page.locator('.card table tbody tr')
		await expect(row).toContainText('https://gone.example.com')
		await row.getByRole('button', {name: 'Unlink'}).click()

		const dialog = page.getByRole('dialog', {name: 'Unlink https://gone.example.com'})
		await dialog.locator('#unlinkIdentityPassword').fill('wrong password')
		await dialog.getByRole('button', {name: 'Unlink'}).click()
		await expect(page.locator('.global-notification')).toContainText('Wrong username or password.')
		await expect(dialog).toBeVisible()
		expect(await listIdentities(apiContext, userToken)).toHaveLength(1)

		await dialog.locator('#unlinkIdentityPassword').fill(TEST_PASSWORD)
		await dialog.getByRole('button', {name: 'Unlink'}).click()

		await expect(page.getByText('You have not linked any accounts yet.')).toBeVisible()
		await page.reload()
		await expect(page.getByText('You have not linked any accounts yet.')).toBeVisible()
		expect(await listIdentities(apiContext, userToken)).toHaveLength(0)
	})
	test('logs a local user with a linked Dex identity into their own account', async ({page, apiContext, currentUser}) => {
		// Not authenticatedPage: its init script would restore the old token after logout.
		await setupApiUrl(page)
		await page.goto('/login')
		await page.locator('input[id=username]').fill(currentUser.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()
		await expect(page).toHaveURL('/')

		await gotoUserSettings(page, 'linked-accounts')
		await linkDexAccount(page)
		await page.locator('#linkIdentityPassword').fill(TEST_PASSWORD)
		await page.getByRole('button', {name: 'Link account'}).click()
		await expect(page.locator('.card table')).toContainText('Dex')

		await page.locator('.navbar .username-dropdown-trigger').click()
		await page.locator('.navbar .dropdown-item').filter({hasText: 'Logout'}).click()
		await expect(page).toHaveURL(/\/login/)

		await page.getByRole('button', {name: 'Log in with Dex'}).click()
		await expect(page.locator('h2')).toContainText('Log in to Your Account')
		await page.locator('#login').fill('test@example.com')
		await page.locator('#password').fill('12345678')
		await page.locator('#submit-login').click()

		await expect(page.locator('main h1')).toContainText(currentUser.username)
		const token = await page.evaluate(() => localStorage.getItem('token'))
		const response = await apiContext.get('user', {
			headers: {Authorization: `Bearer ${token}`},
		})
		expect(response.ok()).toBe(true)
		expect((await response.json()).id).toBe(currentUser.id)
	})
})
