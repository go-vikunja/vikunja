import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'

test.describe('Project webhooks', () => {
	test.beforeEach(async ({authenticatedPage, currentUser}) => {
		await ProjectFactory.create(1, {id: 1, owner_id: currentUser.id}, false)
	})

	test('validates the target URL', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')
		await page.locator('#targetUrl').fill('not-a-url')
		await page.locator('#targetUrl').blur()
		await expect(page.locator('.help.is-danger')).toContainText(/valid URL/i)
	})

	test('requires at least one event', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')
		await page.locator('#targetUrl').fill('https://example.com/hook')
		await page.getByRole('button', {name: /create webhook/i}).click()
		await expect(page.locator('.help.is-danger')).toContainText(/at least one event/i)
	})
})
