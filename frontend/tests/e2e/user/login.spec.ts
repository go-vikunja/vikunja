import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

interface LoginCredentials {
	username: string
	password: string
}

const testAndAssertFailed = async (page: Page, fixture: LoginCredentials): Promise<void> => {
	const loginPromise = page.waitForResponse(response =>
		response.url().includes('/login') && response.request().method() === 'POST',
	)

	await page.goto('/login')
	await page.locator('input[id=username]').fill(fixture.username)
	await page.locator('input[id=password]').fill(fixture.password)
	await page.locator('.button').filter({hasText: 'Login'}).click()

	await loginPromise
	await expect(page).toHaveURL('/login')
	await expect(page.locator('div.message.danger')).toContainText('Wrong username or password.')
}

const credentials: LoginCredentials = {
	username: 'test',
	password: TEST_PASSWORD,
}

async function login(page: Page): Promise<void> {
	await page.locator('input[id=username]').fill(credentials.username)
	await page.locator('input[id=password]').fill(credentials.password)
	await page.locator('.button').filter({hasText: 'Login'}).click()
	await expect(page).toHaveURL('/')
}

test.describe('Login', () => {
	test.beforeEach(async ({page, apiContext}) => {
		await setupApiUrl(page)
		await UserFactory.create(1, {username: credentials.username})
		await page.clock.setFixedTime(new Date(1625656161057)) // 13:00
	})

	test('Should log in with the right credentials', async ({page}) => {
		await page.goto('/login')
		await login(page)
		await expect(page.locator('main h1')).toContainText(credentials.username)
	})

	test('Should fail with a bad password', async ({page}) => {
		const fixture = {
			username: 'test',
			password: '123456',
		}

		await testAndAssertFailed(page, fixture)
	})

	test('Should fail with a bad username', async ({page}) => {
		const fixture = {
			username: 'loremipsum',
			password: TEST_PASSWORD,
		}

		await testAndAssertFailed(page, fixture)
	})

	test('Should redirect to /login when no user is logged in', async ({page}) => {
		await page.goto('/')
		await expect(page).toHaveURL(/\/login/)
	})

	test('Should refresh an expired token without a refresh cookie once and then drop it', async ({page}) => {
		const refreshRequests: string[] = []
		page.on('request', request => {
			if (request.url().includes('/user/token/refresh')) {
				refreshRequests.push(request.url())
			}
		})

		// exp is relative to the fixed clock from beforeEach.
		const payload = Buffer.from(JSON.stringify({
			id: 1,
			type: 1,
			exp: Math.floor(1625656161057 / 1000) - 3600,
		})).toString('base64')
		await page.goto('/login')
		await page.evaluate(token => localStorage.setItem('token', token), `header.${payload}.signature`)

		await page.goto('/')
		await expect(page).toHaveURL(/\/login/)
		await expect(page.locator('input[id=username]')).toBeVisible()
		await expect.poll(() => page.evaluate(() => localStorage.getItem('token'))).toBeNull()
		expect(refreshRequests).toHaveLength(1)
		expect(refreshRequests[0]).toContain('/api/v2/')

		await page.reload()
		await expect(page.locator('input[id=username]')).toBeVisible()
		expect(refreshRequests).toHaveLength(1)
	})

	for (const failure of ['network', '503'] as const) {
		test(`Should recover an expired session after a transient ${failure} refresh failure without reloading`, async ({page}) => {
			await page.goto('/login')
			await login(page)
			await expect(page.locator('main h1')).toContainText(credentials.username)

			const payload = Buffer.from(JSON.stringify({
				id: 1,
				type: 1,
				exp: Math.floor(1625656161057 / 1000) - 3600,
			})).toString('base64')
			const expiredToken = `header.${payload}.signature`
			await page.evaluate(token => localStorage.setItem('token', token), expiredToken)

			let failedRefreshes = 0
			await page.route('**/api/v2/user/token/refresh', async route => {
				failedRefreshes++
				if (failure === 'network') {
					await route.abort('failed')
				} else {
					await route.fulfill({status: 503, json: {status: 503, detail: 'Service unavailable'}})
				}
			})

			await page.reload()
			await expect(page).toHaveURL(/\/login/)
			await expect(page.locator('input[id=username]')).toBeVisible()
			expect(failedRefreshes).toBeGreaterThanOrEqual(2)
			expect(await page.evaluate(() => localStorage.getItem('token'))).toBe(expiredToken)

			await page.unroute('**/api/v2/user/token/refresh')
			const refreshResponse = page.waitForResponse(response =>
				response.url().includes('/api/v2/user/token/refresh') && response.status() === 200,
			)
			await page.locator('.reset-password-link').click()
			await refreshResponse
			await page.getByRole('link', {name: 'Login', exact: true}).click()
			await expect(page).toHaveURL('/')
			await expect(page.locator('main h1')).toContainText(credentials.username)
			expect(await page.evaluate(() => localStorage.getItem('token'))).not.toBe(expiredToken)
		})
	}

	test('Should not show login form inside authenticated app shell after login', async ({page}) => {
		await page.goto('/login')

		// Set up a flag that tracks if login form and navbar are ever visible simultaneously
		await page.evaluate(() => {
			(window as any).__loginFormFlashDetected = false
			const observer = new MutationObserver(() => {
				const hasLoginForm = !!document.querySelector('#loginform')
				const hasNavbar = !!document.querySelector('nav[aria-label="main navigation"]')
				if (hasLoginForm && hasNavbar) {
					(window as any).__loginFormFlashDetected = true
				}
			})
			observer.observe(document.body, {childList: true, subtree: true, attributes: true})
			;(window as any).__loginFormFlashObserver = observer
		})

		await login(page)

		const flashDetected = await page.evaluate(() => {
			(window as any).__loginFormFlashObserver.disconnect()
			return (window as any).__loginFormFlashDetected
		})

		expect(flashDetected).toBe(false)
	})

	test('Should redirect to the previous route after logging in', async ({page}) => {
		const projects = await ProjectFactory.create(1)
		await page.goto(`/projects/${projects[0].id}/1`)

		await expect(page).toHaveURL(/\/login/)

		// Login without expecting redirect to /
		await page.locator('input[id=username]').fill(credentials.username)
		await page.locator('input[id=password]').fill(credentials.password)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		// Should redirect back to the project route
		await expect(page).toHaveURL(new RegExp(`/projects/${projects[0].id}/1`))
	})
})
