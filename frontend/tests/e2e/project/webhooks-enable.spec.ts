import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {WebhookFactory} from '../../factories/webhook'

test('enables a retained webhook and takes it over without changing its destination', async ({
	authenticatedPage: page,
	currentUser,
	apiContext,
	userToken,
}) => {
	await ProjectFactory.create(1, {id: 1, owner_id: currentUser.id}, false)
	const [formerCreator] = await UserFactory.create(1, {id: 2}, false)
	const targetUrl = 'https://example.com/retained-webhook'
	const [webhook] = await WebhookFactory.create(1, {
		project_id: 1,
		target_url: targetUrl,
		created_by_id: formerCreator.id,
		is_disabled: true,
	})

	await page.goto('/projects/1/settings/webhooks')
	const row = page.getByRole('row').filter({hasText: targetUrl})
	await expect(row).toContainText('Disabled.')
	await expect(row).toContainText('your project access')
	let release!: () => void
	let received!: () => void
	const responseGate = new Promise<void>(resolve => { release = resolve })
	const requestReceived = new Promise<void>(resolve => { received = resolve })
	await page.route(`**/api/v2/projects/1/webhooks/${webhook.id}`, async route => {
		if (route.request().method() !== 'PUT') return route.continue()
		const response = await route.fetch()
		received()
		await responseGate
		await route.fulfill({response})
	})
	const enableButton = row.getByRole('button', {name: 'Enable webhook', exact: true})
	await enableButton.focus()
	await enableButton.press('Enter')
	await requestReceived
	await expect(enableButton).toBeFocused()
	await expect(enableButton).toHaveAttribute('aria-disabled', 'true')
	release()
	await expect(enableButton).toHaveCount(0)
	await expect(row.getByRole('button', {name: 'Delete this webhook', exact: true})).toBeFocused()

	await page.reload()
	await expect(row).toBeVisible()
	await expect(row).not.toContainText('Disabled.')
	await expect(row).toContainText(currentUser.username)
	const response = await apiContext.get('/api/v2/projects/1/webhooks', {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(response.ok()).toBe(true)
	expect((await response.json()).items).toEqual([
		expect.objectContaining({
			id: webhook.id,
			is_disabled: false,
			target_url: targetUrl,
			events: ['task.created'],
			created_by: expect.objectContaining({id: currentUser.id}),
		}),
	])
})
