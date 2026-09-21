import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {LicenseFactory} from '../../factories/license'
import {login} from '../../support/authenticateUser'

test.beforeEach(async () => { await LicenseFactory.enable(['admin_panel']) })
test.afterEach(async () => { await LicenseFactory.disable() })

test('admin password changes are usable after reload', async ({page, apiContext}) => {
	const [admin] = await UserFactory.create(1, {id: 1, is_admin: true}, false)
	const [target] = await UserFactory.create(1, {id: 2, username: 'password-target'}, false)
	await login(page, apiContext, admin)
	await page.goto('/admin/users')
	await page.getByRole('row', {name: /password-target/}).getByRole('button', {name: 'Details', exact: true}).click()
	const dialog = page.getByRole('dialog')
	await dialog.getByLabel('New password', {exact: true}).fill('updated-admin-password')
	const changed = page.waitForResponse(r => r.url().endsWith('/admin/users/2/password') && r.request().method() === 'PATCH')
	await dialog.getByRole('button', {name: 'Set password', exact: true}).click()
	expect((await changed).ok()).toBe(true)
	await page.reload()
	const signedIn = await apiContext.post('/api/v2/login', {data: {username: target.username, password: 'updated-admin-password'}})
	expect(signedIn.ok()).toBe(true)
	expect((await signedIn.json()).token).toBeTruthy()
})

test('admin project owner reassignment persists after reload', async ({page, apiContext}) => {
	const [admin] = await UserFactory.create(1, {id: 1, is_admin: true}, false)
	const [owner] = await UserFactory.create(1, {id: 2, username: 'transfer-owner'}, false)
	const [project] = await ProjectFactory.create(1, {owner_id: admin.id, title: 'Transfer project'}, false)
	const {token} = await login(page, apiContext, admin)
	await page.goto('/admin/projects')
	const row = page.getByRole('row', {name: /Transfer project/})
	await row.locator('.dropdown-trigger').click()
	await page.getByText('Reassign owner', {exact: true}).click()
	const dialog = page.getByRole('dialog')
	const search = dialog.getByRole('combobox')
	await search.pressSequentially('transfer-owner')
	await page.getByRole('option', {name: /^transfer-owner/}).click()
	await dialog.getByRole('button', {name: 'Reassign owner', exact: true}).click()
	await expect(row).toContainText('transfer-owner')
	await page.reload()
	await expect(row).toContainText('transfer-owner')
	const stored = await apiContext.get('/api/v2/admin/projects', {headers: {Authorization: `Bearer ${token}`}})
	expect((await stored.json()).items).toContainEqual(expect.objectContaining({id: project.id, owner: expect.objectContaining({id: owner.id})}))
})
