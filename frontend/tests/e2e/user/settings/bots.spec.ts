import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {serverPageSize} from '../../../support/pagination'

test('manages a bot and its scoped token across reloads', async ({authenticatedPage: page, apiContext, userToken}) => {
	await gotoUserSettings(page, 'bots')
	await page.getByPlaceholder('bot-myassistant').fill('assistant')
	await page.getByPlaceholder('My Assistant').fill('Original name')
	await page.getByRole('button', {name: 'Create bot', exact: true}).click()
	const card = page.locator('.bot-card')
	await expect(card).toContainText('bot-assistant')
	await card.getByRole('button', {name: 'Edit', exact: true}).click()
	await card.locator('.bot-name-input').fill('Renamed bot')
	await card.getByRole('button', {name: 'Save', exact: true}).click()
	await expect(card).toContainText('Renamed bot')
	await card.getByRole('button', {name: 'Disable', exact: true}).click()
	await expect(card.locator('.status')).toHaveText('Disabled')
	await page.reload()
	await expect(card).toContainText('Renamed bot')
	await expect(card.locator('.status')).toHaveText('Disabled')
	await expect(card.getByRole('button', {name: 'Create token', exact: true})).toHaveCount(0)
	await expect(card).toContainText('API tokens of a disabled bot are rejected')
	await card.getByRole('button', {name: 'Enable', exact: true}).click()
	await expect(card.locator('.status')).toHaveText('Active')
	await page.reload()
	await expect(card.locator('.status')).toHaveText('Active')
	await card.getByRole('button', {name: 'Create token', exact: true}).click()
	await card.locator('#apiTokenTitle').fill('Bot token')
	await card.getByRole('button', {name: 'Read only', exact: true}).click()
	const created = page.waitForResponse(r => r.url().endsWith('/api/v2/tokens') && r.request().method() === 'POST')
	await card.locator('form').getByRole('button', {name: 'Create token', exact: true}).click()
	const token = await (await created).json()
	await expect(card.locator('code')).toHaveText(token.token)
	await page.reload()
	await expect(card.locator('tbody')).toContainText('Bot token')
	await expect(card.locator('code')).toHaveCount(0)
	await card.locator('tbody').getByRole('button', {name: 'Delete', exact: true}).click()
	await expect(card.locator('tbody tr')).toHaveCount(0)
	await card.locator('.bot-actions').getByRole('button', {name: 'Delete', exact: true}).click()
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(card).toHaveCount(0)
	await page.reload()
	await expect(card).toHaveCount(0)
	const response = await apiContext.get('/api/v2/user/bots', {headers: {Authorization: `Bearer ${userToken}`}})
	expect(response.ok()).toBe(true)
	expect((await response.json()).items).toEqual([])
})

test('pages through bots and leaves an emptied last page', async ({authenticatedPage: page, apiContext, userToken}) => {
	const pageSize = await serverPageSize(apiContext)
	for (let i = 1; i <= pageSize + 1; i++) {
		const response = await apiContext.post('/api/v2/user/bots', {
			headers: {Authorization: `Bearer ${userToken}`},
			data: {username: `bot-paged-${i}`},
		})
		expect(response.ok()).toBe(true)
	}
	await gotoUserSettings(page, 'bots')
	const cards = page.locator('.bot-card')
	await expect(cards).toHaveCount(pageSize)
	await page.getByRole('link', {name: 'Goto page 2'}).click()
	await expect(page).toHaveURL(/[?&]page=2/)
	await expect(cards).toHaveCount(1)
	await cards.locator('.bot-actions').getByRole('button', {name: 'Delete', exact: true}).click()
	await page.locator('[data-cy="modalPrimary"]').click()
	await expect(page).not.toHaveURL(/[?&]page=2/)
	await expect(cards).toHaveCount(pageSize)
	await expect(page.getByRole('link', {name: 'Goto page 2'})).toHaveCount(0)
})

test('shows a bot created past the first page on its page', async ({authenticatedPage: page, apiContext, userToken}) => {
	const pageSize = await serverPageSize(apiContext)
	for (let i = 1; i <= pageSize; i++) {
		const response = await apiContext.post('/api/v2/user/bots', {
			headers: {Authorization: `Bearer ${userToken}`},
			data: {username: `bot-full-${i}`},
		})
		expect(response.ok()).toBe(true)
	}
	await gotoUserSettings(page, 'bots')
	const cards = page.locator('.bot-card')
	await expect(cards).toHaveCount(pageSize)
	await page.getByRole('button', {name: 'Create bot', exact: true}).click()
	await page.getByPlaceholder('bot-myassistant').fill('overflow')
	await page.getByRole('button', {name: 'Create bot', exact: true}).click()
	await expect(page.locator('.global-notification')).toContainText('bot-overflow')
	await expect(page).toHaveURL(/[?&]page=2/)
	await expect(cards).toHaveCount(1)
	await expect(cards).toContainText('bot-overflow')
})
