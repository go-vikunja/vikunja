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
			// Router guard redirects to not-found.
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

			// Guard redirects to not-found; the admin shell should not render.
			await expect(page.locator('.side-nav-shell')).not.toBeVisible()
		})
	})
})
