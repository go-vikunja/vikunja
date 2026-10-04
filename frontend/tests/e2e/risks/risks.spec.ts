import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {RiskFactory} from '../../factories/risk'

test.describe('Risks', () => {
	test.beforeEach(async ({currentUser}) => {
		await ProjectFactory.create(1, {id: 1, title: 'Risk project', owner_id: currentUser.id}, false)
		await RiskFactory.truncate()
	})

	test('has a Risks entry in the sidebar that opens the register', async ({authenticatedPage: page}) => {
		await RiskFactory.create(1, {project_id: 1, title: 'Supplier delay', probability: 5, impact: 5}, false)

		await page.goto('/')
		await page.locator('.menu-container').getByRole('link', {name: 'Risks'}).click()

		await expect(page).toHaveURL(/\/risks$/)
		await expect(page.getByRole('heading', {name: 'Risk register'})).toBeVisible()
		const row = page.locator('table.table tbody tr', {hasText: 'Supplier delay'})
		await expect(row).toBeVisible()
		// 5 x 5 is the highest score.
		await expect(row).toContainText('Critical')
		await expect(row).toContainText('25')
	})

	test('creates a risk on the project page and shows its score', async ({authenticatedPage: page}) => {
		await page.goto('/projects/1/risks')
		await page.getByRole('button', {name: 'Add risk'}).click()

		const dialog = page.locator('dialog[open]')
		await dialog.getByLabel('Title').fill('Key person leaves')
		await dialog.getByLabel('Category').fill('People')
		await dialog.getByLabel('Probability').selectOption({label: '4 - Likely'})
		await dialog.getByLabel('Impact').selectOption({label: '4 - Major'})
		// 4 x 4 = 16 is still high, one below critical.
		await expect(dialog).toContainText('Score 16, High')

		const created = page.waitForResponse(r =>
			r.url().includes('/projects/1/risks') && r.request().method() === 'POST',
		)
		await dialog.getByRole('button', {name: 'Save'}).click()
		expect((await created).status()).toBe(201)

		const row = page.locator('table.table tbody tr', {hasText: 'Key person leaves'})
		await expect(row).toBeVisible()
		await expect(row).toContainText('People')
		await expect(row).toContainText('High')
	})

	test('closes and reopens a risk, and keeps closed risks out of the list until asked', async ({authenticatedPage: page}) => {
		await RiskFactory.create(1, {project_id: 1, title: 'Late permit'}, false)
		await page.goto('/projects/1/risks')

		await page.locator('table.table tbody tr', {hasText: 'Late permit'}).click()
		const detail = page.locator('dialog[open]')
		await detail.getByRole('button', {name: 'Close risk'}).click()

		// The status dialog opens on top of the detail.
		const status = page.locator('dialog[open]').last()
		await status.getByLabel('How was it resolved? (optional)').fill('Permit granted')
		const closed = page.waitForResponse(r =>
			r.url().match(/\/risks\/\d+\/status/) !== null && r.request().method() === 'POST',
		)
		await status.getByRole('button', {name: 'Close risk'}).click()
		expect((await closed).status()).toBe(200)

		// Closed risks are hidden by default.
		await page.locator('dialog[open]').getByRole('button', {name: 'Close'}).first().click()
		await expect(page.locator('table.table tbody tr', {hasText: 'Late permit'})).toHaveCount(0)

		await page.getByLabel('Show closed').check()
		const row = page.locator('table.table tbody tr', {hasText: 'Late permit'})
		await expect(row).toBeVisible()
		await expect(row).toContainText('Closed')

		// Reopening is one click.
		await row.click()
		const reopened = page.waitForResponse(r =>
			r.url().match(/\/risks\/\d+\/status/) !== null && r.request().method() === 'POST',
		)
		await page.locator('dialog[open]').getByRole('button', {name: 'Reopen'}).click()
		expect((await reopened).status()).toBe(200)
		await expect(page.locator('dialog[open]')).toContainText('Open')
	})

	test('narrows the register with a filter from the URL', async ({authenticatedPage: page}) => {
		await RiskFactory.create(1, {project_id: 1, title: 'Minor glitch', probability: 1, impact: 1}, false)
		await RiskFactory.create(1, {id: 2, project_id: 1, title: 'Total outage', probability: 5, impact: 5}, false)

		await page.goto('/risks?rating=critical')

		await expect(page.locator('table.table tbody tr', {hasText: 'Total outage'})).toBeVisible()
		await expect(page.locator('table.table tbody tr', {hasText: 'Minor glitch'})).toHaveCount(0)
	})

	test('offers the Excel and PDF export', async ({authenticatedPage: page}) => {
		await page.goto('/risks')
		await page.getByRole('button', {name: 'Export risks'}).click()

		await expect(page.getByRole('button', {name: 'Excel (.xlsx)'})).toBeVisible()
		await expect(page.getByRole('button', {name: 'PDF (print view)'})).toBeVisible()
	})
})
