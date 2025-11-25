import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'

async function logout(page) {
	await page.locator('.navbar .username-dropdown-trigger').click()
	await page.locator('.navbar .dropdown-item').filter({hasText: 'Logout'}).click()
}

test.describe('Log out', () => {
	test.use({
		// All tests in this describe block use the authenticatedPage fixture
	})

	test('Logs the user out', async ({authenticatedPage: page}) => {
		await page.goto('/')

		// Check that token exists before logout
		const tokenBefore = await page.evaluate(() => localStorage.getItem('token'))
		expect(tokenBefore).not.toBeNull()

		await logout(page)

		// Check URL redirects to login
		await expect(page).toHaveURL(/\/login/)

		// Check that token is removed after logout
		const tokenAfter = await page.evaluate(() => localStorage.getItem('token'))
		expect(tokenAfter).toBeNull()
	})

	test.skip('Should clear the project history after logging the user out', async ({authenticatedPage: page, apiContext}) => {
		const projects = await ProjectFactory.create(1)
		await page.goto(`/projects/${projects[0].id}`)

		// Check that project history exists
		const historyBefore = await page.evaluate(() => localStorage.getItem('projectHistory'))
		expect(historyBefore).not.toBeNull()

		await logout(page)

		// Wait a bit to make re-loading of the project and associated entities visible
		await page.waitForTimeout(1000)

		// Check URL redirects to login
		await expect(page).toHaveURL(/\/login/)

		// Check that project history is cleared
		const historyAfter = await page.evaluate(() => localStorage.getItem('projectHistory'))
		expect(historyAfter).toBeNull()
	})
})
