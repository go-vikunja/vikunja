import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {LicenseFactory} from '../../factories/license'
import {login} from '../../support/authenticateUser'

test.describe('Admin panel', () => {
	test.describe('with admin_panel feature licensed', () => {
		test.beforeEach(async () => {
			await LicenseFactory.enable(['admin_panel'])
		})

		test.afterEach(async () => {
			await LicenseFactory.disable()
		})

		test('an admin user can open /admin and see the overview', async ({page, apiContext}) => {
			const [admin] = await UserFactory.create(1, {is_admin: true}, false)
			await login(page, apiContext, admin)

			await page.goto('/admin')

			await expect(page.locator('.side-nav-shell > nav')).toBeVisible()
			await expect(page.locator('.card-header-title', {hasText: 'Overview'})).toBeVisible()
			await expect(page.locator('.admin-overview__card').first()).toBeVisible()
		})

		test('a non-admin user visiting /admin lands on the not-found page', async ({authenticatedPage: page}) => {
			await page.goto('/admin')
			await expect(page).not.toHaveURL(/\/admin$/)
		})

		test('an admin can navigate to users and projects tabs', async ({page, apiContext}) => {
			const [admin] = await UserFactory.create(1, {is_admin: true}, false)
			await login(page, apiContext, admin)
			await page.goto('/admin')

			const nav = page.locator('.side-nav-shell > nav')
			await nav.getByRole('link', {name: /users/i}).click()
			await expect(page).toHaveURL(/\/admin\/users/)

			await nav.getByRole('link', {name: /projects/i}).click()
			await expect(page).toHaveURL(/\/admin\/projects/)
		})
	})

	test.describe('without license', () => {
		test.beforeEach(async () => {
			await LicenseFactory.disable()
		})

		test('even an admin user cannot access /admin', async ({page, apiContext}) => {
			const [admin] = await UserFactory.create(1, {is_admin: true}, false)
			await login(page, apiContext, admin)

			await page.goto('/admin')

			await expect(page.locator('.side-nav-shell')).not.toBeVisible()
		})
	})
})


test('admin user changes persist and license removal rejects new requests', async ({page, apiContext}) => {
	await LicenseFactory.enable(['admin_panel'])
	const [admin] = await UserFactory.create(1, {is_admin: true}, false)
	const {token} = await login(page, apiContext, admin)
	await page.goto('/admin/users')
	await page.getByRole('button', {name: 'Add user', exact: true}).click()
	const dialog = page.locator('dialog[open]')
	await dialog.getByLabel('Username', {exact: true}).fill('managed-user')
	await dialog.getByLabel('Email address', {exact: true}).fill('managed@example.com')
	await dialog.getByLabel('Password', {exact: true}).fill('new-password-123')
	await dialog.getByRole('button', {name: 'Create user', exact: true}).click()
	const row = page.locator('tbody tr', {hasText: 'managed-user'})
	await expect(row).toBeVisible()
	await page.reload()
	await row.getByRole('button', {name: 'Details', exact: true}).click()
	await dialog.locator('select').selectOption('2')
	await dialog.getByRole('button', {name: 'Save changes', exact: true}).click()
	await expect(row).toContainText('Disabled')
	await page.reload()
	await expect(row).toContainText('Disabled')
	await row.getByRole('button', {name: 'Details', exact: true}).click()
	await dialog.getByRole('button', {name: 'Delete', exact: true}).click()
	await dialog.getByRole('button', {name: 'Delete now', exact: true}).click()
	await expect(row).toHaveCount(0)
	await page.reload()
	await expect(row).toHaveCount(0)
	const headers = {Authorization: `Bearer ${token}`}
	const users = await apiContext.get('/api/v2/admin/users?q=managed-user', {headers})
	expect(users.ok()).toBe(true)
	expect((await users.json()).items ?? []).toEqual([])
	await LicenseFactory.disable()
	const blocked = await apiContext.get('/api/v2/admin/users', {headers})
	expect(blocked.status()).toBe(404)
})
