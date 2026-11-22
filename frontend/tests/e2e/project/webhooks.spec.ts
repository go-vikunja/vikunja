import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {WebhookFactory} from '../../factories/webhook'
import {serverPageSize} from '../../support/pagination'

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
			r.url().includes('/projects/1/webhooks') && r.request().method() === 'POST',
		)
		await page.getByRole('button', {name: /create webhook/i}).click()
		await created

		const row = page.locator('table.table tbody tr', {hasText: 'example.com/hook'})
		await expect(row).toBeVisible()
		await page.reload()
		await expect(row).toBeVisible()

		const deleted = page.waitForResponse(r =>
			r.url().match(/\/projects\/1\/webhooks\/\d+/) !== null && r.request().method() === 'DELETE',
		)
		await row.locator('.button.is-danger').click()
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await deleted

		await expect(row).toHaveCount(0)
		await page.reload()
		await expect(row).toHaveCount(0)
	})

	test('selects all events at once', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')

		const eventChecks = page.locator('.available-events-check input')
		const selectAll = page.locator('.webhook-events-select-all input')
		await expect(eventChecks.first()).toBeAttached()

		await page.locator('.webhook-events-select-all .base-checkbox__label').click()
		for (const check of await eventChecks.all()) {
			await expect(check).toBeChecked()
		}

		await page.locator('.available-events-check', {hasText: 'task.created'})
			.locator('.base-checkbox__label').click()
		await expect(selectAll).not.toBeChecked()

		await page.locator('.available-events-check', {hasText: 'task.created'})
			.locator('.base-checkbox__label').click()
		await expect(selectAll).toBeChecked()

		await page.locator('.webhook-events-select-all .base-checkbox__label').click()
		for (const check of await eventChecks.all()) {
			await expect(check).not.toBeChecked()
		}

		await page.locator('#targetUrl').fill('https://example.com/all')
		await page.locator('.webhook-events-select-all .base-checkbox__label').click()
		const events = await page.locator('.available-events-check .fancy-checkbox__content').allTextContents()
		await page.getByRole('button', {name: /create webhook/i}).click()

		const row = page.locator('table.table tbody tr', {hasText: 'example.com/all'})
		await expect(row.locator('td').nth(1)).toHaveText(events.map(e => e.trim()).join(', '))
	})

	test('selects all events of a group', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')

		const projectGroup = page.locator('.webhook-events-group', {hasText: 'project.deleted'})
		await projectGroup.locator('.webhook-events-group-toggle .base-checkbox__label').click()

		for (const check of await projectGroup.locator('.available-events-check input').all()) {
			await expect(check).toBeChecked()
		}
		await expect(page.locator('.available-events-check', {hasText: 'task.created'}).locator('input')).not.toBeChecked()
		await expect(page.locator('.webhook-events-select-all input')).not.toBeChecked()
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

	test('pages the list and deletes a webhook on a later page', async ({authenticatedPage: page, currentUser, apiContext}) => {
		const pageSize = await serverPageSize(apiContext)
		await WebhookFactory.create(pageSize + 1, {
			project_id: 1,
			target_url: i => `https://example.com/hook-${i}`,
			created_by_id: currentUser.id,
		})

		await page.goto('/projects/1/settings/webhooks')
		await page.waitForLoadState('networkidle')

		const rows = page.locator('table.table tbody tr')
		await expect(rows).toHaveCount(pageSize)

		const secondPage = page.waitForResponse(r =>
			r.url().includes('/projects/1/webhooks') && r.url().includes('page=2') && r.request().method() === 'GET',
		)
		await page.getByRole('button', {name: 'Goto page 2'}).click()
		await secondPage
		await expect(rows).toHaveCount(1)

		const deleted = page.waitForResponse(r =>
			r.url().match(/\/projects\/1\/webhooks\/\d+/) !== null && r.request().method() === 'DELETE',
		)
		await rows.first().locator('.button.is-danger').click()
		await page.locator('dialog[open] .modal-content .actions .button').filter({hasText: 'Do it!'}).click()
		await deleted

		await expect(rows).toHaveCount(pageSize)
		await expect(page.locator('nav.pagination')).toHaveCount(0)
		await page.reload()
		await expect(rows).toHaveCount(pageSize)
		await expect(page.locator('nav.pagination')).toHaveCount(0)
	})
})
