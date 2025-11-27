import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
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
	test.beforeEach(async ({apiContext}) => {
		await UserFactory.create(1, {username: credentials.username})
	})

	test('Should log in with the right credentials', async ({page}) => {
		await page.goto('/login')
		await login(page)
		await page.clock.install({time: new Date(1625656161057)}) // 13:00
		// Use more specific selector to avoid strict mode violation
		await expect(page.locator('main h2')).toContainText(`Hi ${credentials.username}!`)
	})

	// FIXME: request timeout for the request that's awaited
	test.skip('Should fail with a bad password', async ({page}) => {
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

	// FIXME: request timeout
	test.skip('Should redirect to the previous route after logging in', async ({page}) => {
		const projects = await ProjectFactory.create(1)
		await page.goto(`/projects/${projects[0].id}/1`)

		await expect(page).toHaveURL(/\/login/)

		await login(page)

		await expect(page).toHaveURL(new RegExp(`/projects/${projects[0].id}/1`))
	})
})
