import {test, expect} from '../../support/fixtures'

test('CSV detection, preview and import persist after reload', async ({authenticatedPage: page, apiContext, userToken}) => {
	await page.goto('/migrate/csv')
	await page.locator('input[type=file]').setInputFiles({name: 'tasks.csv', mimeType: 'text/csv', buffer: Buffer.from('title,description,project\nImported CSV task,Stored description,CSV project\n')})
	await expect(page.locator('.preview-tasks')).toContainText('Imported CSV task')
	await page.getByRole('button', {name: 'Import Tasks', exact: true}).click()
	await expect(page.locator('.success-step [role=status]')).toContainText(/successfully|finished/i, {timeout: 20000})
	await page.goto('/')
	await page.reload()
	await expect(page.getByRole('link', {name: 'CSV project', exact: true}).first()).toContainText('CSV project')
	const stored = await apiContext.get('/api/v2/tasks', {headers: {Authorization: `Bearer ${userToken}`}})
	expect(stored.ok()).toBe(true)
	expect((await stored.json()).items).toContainEqual(expect.objectContaining({title: 'Imported CSV task', description: 'Stored description'}))
})
