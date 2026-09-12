import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TeamFactory} from '../../factories/team'
import {LicenseFactory} from '../../factories/license'
import {login, setupApiUrl} from '../../support/authenticateUser'

test.describe('Invite links', () => {
	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
		await LicenseFactory.enable(['admin_panel', 'user_invites'])
	})

	test.afterEach(async () => {
		await LicenseFactory.disable()
	})

	test('admin creates and copies a link, guest registers and joins its team', async ({page, apiContext, browser, baseURL}) => {
		const [admin] = await UserFactory.create(1, {is_admin: true, email: 'admin@example.com', name: 'Invite Admin'}, false)
		const [team] = await TeamFactory.create(1, {id: 1, name: 'Invited team'}, false)
		await login(page, apiContext, admin)
		await page.context().grantPermissions(['clipboard-read', 'clipboard-write'])
		await page.goto('/admin')
		await page.getByRole('link', {name: 'Invite links', exact: true}).click()
		await page.getByRole('button', {name: 'Create invite link', exact: true}).click()
		await expect(page.getByRole('dialog')).toHaveAccessibleName('Create invite link')
		const skipConfirmation = page.getByRole('checkbox', {name: 'Skip email confirmation'})
		await expect(skipConfirmation).not.toBeChecked()
		await skipConfirmation.check()
		await page.getByRole('dialog').getByRole('button', {name: 'Cancel', exact: true}).click()
		await page.getByRole('button', {name: 'Create invite link', exact: true}).click()
		await expect(skipConfirmation).not.toBeChecked()
		await skipConfirmation.check()
		await page.getByLabel('Link name', {exact: true}).fill('Welcome aboard')
		await page.getByRole('combobox', {name: 'Teams to join'}).click()
		await expect(page.getByRole('listbox')).toHaveAccessibleName('Teams to join')
		await page.getByRole('option', {name: 'Invited team'}).click()
		await page.getByLabel('Maximum uses').fill('1')
		await page.getByRole('button', {name: 'Create invite link', exact: true}).last().click()
		await expect(page.getByText('This link is shown only once.', {exact: false})).toBeVisible()
		const url = await page.getByLabel('Invite link', {exact: true}).inputValue()
		expect(new URL(url).pathname).toBe('/register')
		expect(new URL(url).hash).toMatch(/^#invite-link=.+/)
		await page.getByRole('button', {name: 'Copy link', exact: true}).click()
		await expect.poll(() => page.evaluate(() => navigator.clipboard.readText())).toBe(url)
		await page.getByRole('button', {name: 'Close', exact: true}).last().click()
		await page.reload()
		await expect(page.getByLabel('Invite link', {exact: true})).not.toBeVisible()
		const inviteRow = page.getByRole('row').filter({hasText: 'Welcome aboard'})
		await expect(inviteRow).toBeVisible()
		const creator = inviteRow.getByRole('cell').nth(4)
		await expect(creator).toHaveText('Invite Admin')
		await expect(creator.locator('img')).toBeVisible()
		await expect.poll(() => creator.locator('img').evaluate((img: HTMLImageElement) => img.complete && img.naturalWidth > 0)).toBe(true)
		await page.evaluate(() => localStorage.removeItem('token'))
		await page.context().clearCookies()

		const guest = await browser.newContext({baseURL})
		try {
			const guestPage = await guest.newPage()
			await setupApiUrl(guestPage)
			// The harness serves the frontend and API on separate ports.
			await guestPage.goto(new URL(url).pathname + new URL(url).hash)
			await expect(guestPage.getByText('You will join: Invited team')).toBeVisible()
			await expect(guestPage).toHaveURL('/register')
			await guestPage.getByLabel('Username', {exact: true}).fill('invited-guest')
			await guestPage.getByLabel('Email address', {exact: true}).fill('admin@example.com')
			await guestPage.locator('#password').fill('12345678')
			await guestPage.locator('#register-submit').click()
			await expect(guestPage.getByText('A user with this email address already exists.', {exact: true})).toBeVisible()
			await guestPage.getByLabel('Email address', {exact: true}).fill('invited-guest@example.com')
			await guestPage.locator('#register-submit').click()
			await expect(guestPage).toHaveURL('/')
			const token = await guestPage.evaluate(() => localStorage.getItem('token'))
			const teams = await apiContext.get('teams', {headers: {Authorization: `Bearer ${token}`}})
			expect(teams.ok()).toBeTruthy()
			expect((await teams.json()).map((entry: {id: number}) => entry.id)).toContain(team.id)
			const deadPage = await guest.newPage()
			await setupApiUrl(deadPage)
			await deadPage.goto('/login')
			await deadPage.evaluate(() => localStorage.removeItem('token'))
			await guest.clearCookies()
			await deadPage.goto(new URL(url).pathname + new URL(url).hash)
			await expect(deadPage.getByText('This invite link is invalid or expired.')).toBeVisible()
			await expect(deadPage.locator('#registerform')).not.toBeVisible()
		} finally {
			await guest.close()
		}
		await login(page, apiContext, admin)
		await page.goto('/admin/invite-links')
		await expect(inviteRow.getByRole('cell').nth(2)).toHaveText('1 / 1')
		await inviteRow.getByRole('button', {name: 'Delete', exact: true}).click()
		await page.getByRole('dialog').getByRole('button', {name: 'Do it!', exact: true}).click()
		await expect(inviteRow).not.toBeVisible()
	})

	test('unknown invitation is reachable anonymously and hides registration', async ({page}) => {
		await page.goto('/register#invite-link=unknown-link')
		await expect(page).toHaveURL('/register')
		await expect(page.getByText('This invite link is invalid or expired.')).toBeVisible()
		await expect(page.locator('#registerform')).not.toBeVisible()
		await expect(page.getByRole('link', {name: 'Login', exact: true})).toBeVisible()
	})

	test('changing the fragment revalidates the mounted form and ignores an older response', async ({page, apiContext}) => {
		const [admin] = await UserFactory.create(1, {is_admin: true}, false)
		const teams = await TeamFactory.create(2, {id: '{increment}', name: (id: number) => `Team ${id}`}, false)
		const {token: adminToken} = await login(null, apiContext, admin)
		const links: {token: string}[] = []
		for (const team of teams) {
			const response = await apiContext.post('/api/v2/admin/invite-links', {
				headers: {Authorization: `Bearer ${adminToken}`},
				data: {name: `Invite for ${team.name}`, team_ids: [team.id], skip_email_confirm: true},
			})
			expect(response.status()).toBe(201)
			links.push(await response.json())
		}
		const [first, second] = links
		let releaseFirst!: () => void
		let receivedFirst!: () => void
		const firstReceived = new Promise<void>(resolve => { receivedFirst = resolve })
		const firstReleased = new Promise<void>(resolve => { releaseFirst = resolve })
		await page.route('**/api/v2/invite-links/check', async route => {
			const {token} = route.request().postDataJSON()
			if (token === first.token) {
				receivedFirst()
				await firstReleased
			}
			await route.continue()
		})
		await page.goto(`/register#invite-link=${first.token}`)
		await firstReceived
		await page.evaluate(token => { window.location.hash = `invite-link=${token}` }, second.token)
		await expect(page.getByText('You will join: Team 2')).toBeVisible()
		await expect(page).toHaveURL('/register')
		const firstResponse = page.waitForResponse(response => response.url().endsWith('/invite-links/check') && response.request().postDataJSON().token === first.token)
		releaseFirst()
		expect(await (await firstResponse).json()).toMatchObject({teams: [{id: teams[0].id, name: 'Team 1'}]})
		await expect(page.getByText('You will join: Team 2')).toBeVisible()
	})

	test('tab and direct route are unavailable without user_invites', async ({page, apiContext}) => {
		await LicenseFactory.enable(['admin_panel'])
		const [admin] = await UserFactory.create(1, {is_admin: true}, false)
		await login(page, apiContext, admin)
		await page.goto('/admin')
		await expect(page.getByRole('link', {name: 'Invite links', exact: true})).not.toBeVisible()
		await page.goto('/admin/invite-links')
		await expect(page.getByRole('heading', {name: 'Not found', exact: true})).toBeVisible()
		await expect(page.getByRole('button', {name: 'Create invite link', exact: true})).not.toBeVisible()
	})
})
