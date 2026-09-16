import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {WebhookFactory} from '../../factories/webhook'

test.describe('Project webhooks', () => {
	test.beforeEach(async ({currentUser}) => {
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

	test('creates and deletes a webhook', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')

		await page.locator('#targetUrl').fill('https://example.com/hook')
		await page.locator('.available-events-check', {hasText: 'task.created'})
			.locator('.base-checkbox__label').click()

		const created = page.waitForResponse(r =>
			r.url().includes('/projects/1/webhooks') && r.request().method() === 'PUT',
		)
		await page.getByRole('button', {name: /create webhook/i}).click()
		await created

		const row = page.locator('table.table tbody tr', {hasText: 'example.com/hook'})
		await expect(row).toBeVisible()

		const deleted = page.waitForResponse(r =>
			r.url().match(/\/projects\/1\/webhooks\/\d+/) !== null && r.request().method() === 'DELETE',
		)
		await row.locator('.button.is-danger').click()
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await deleted

		await expect(row).toHaveCount(0)
	})

	test('does not overflow the table with a long target URL', async ({authenticatedPage: page, currentUser}) => {
		const longUrl = 'https://discord.com/api/webhooks/1234567890123456789/' +
			'aVeryLongDiscordWebhookTokenWithoutAnyBreakOpportunitiesAtAllWhatsoever1234567890'
		await WebhookFactory.create(1, {
			project_id: 1,
			target_url: longUrl,
			created_by_id: currentUser.id,
		})

		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')

		const table = page.locator('table.table')
		await expect(table).toContainText('discord.com')

		// The table grows past its container instead of wrapping when the URL cell can't break
		const overflow = await table.evaluate(el => el.getBoundingClientRect().width - el.parentElement!.clientWidth)
		expect(overflow).toBeLessThanOrEqual(1)
	})
})
