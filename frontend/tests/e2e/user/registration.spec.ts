// This test assumes no mailer is set up and all users are activated immediately.

import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'

test.describe('Registration', () => {
	test.beforeEach(async ({page, apiContext}) => {
		await UserFactory.create(1, {
			username: 'test',
		})
		await page.goto('/')
		await page.evaluate(() => localStorage.removeItem('token'))
	})

	test('Should work without issues', async ({page, apiContext}) => {
		const fixture = {
			username: 'testuser',
			password: '12345678',
			email: 'testuser@example.com',
		}

		await page.goto('/register')
		await page.locator('#username').fill(fixture.username)
		await page.locator('#email').fill(fixture.email)
		await page.locator('#password').fill(fixture.password)
		await page.locator('#register-submit').click()
		await expect(page).toHaveURL(/\//)
		await page.clock.install({time: new Date(1625656161057)}) // 13:00
		await expect(page.locator('main h2')).toContainText(`Hi ${fixture.username}!`)
	})

	test('Should fail', async ({page, apiContext}) => {
		const fixture = {
			username: 'test',
			password: '12345678',
			email: 'testuser@example.com',
		}

		await page.goto('/register')
		await page.locator('#username').fill(fixture.username)
		await page.locator('#email').fill(fixture.email)
		await page.locator('#password').fill(fixture.password)
		await page.locator('#register-submit').click()
		await expect(page.locator('div.message.danger')).toContainText('A user with this username already exists.')
	})
})
