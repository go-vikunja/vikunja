import type {APIRequestContext, Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {createDefaultViews} from '../project/prepareProjects'
import {login, setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

async function seedUser() {
	const [user] = await UserFactory.create(1, {default_project_id: 1})
	const [project] = await ProjectFactory.create(1, {id: 1, owner_id: user.id, title: 'Inbox'})
	await createDefaultViews(project.id)
	return {user, project}
}

function prefillUrl(params: Record<string, string>) {
	return '/?' + new URLSearchParams(params).toString()
}

async function createdTask(page: Page, apiContext: APIRequestContext, token: string) {
	await expect(page).toHaveURL(/\/tasks\/\d+$/)
	const id = page.url().match(/\/tasks\/(\d+)$/)![1]
	const response = await apiContext.get(`/api/v2/tasks/${id}`, {headers: {Authorization: `Bearer ${token}`}})
	expect(response.ok()).toBe(true)
	return response.json()
}

test.describe('Quick add prefill', () => {
	test('Creates a task with the page title and a link to the page', async ({page, apiContext}) => {
		const {user, project} = await seedUser()
		const {token} = await login(page, apiContext, user)

		await page.goto(prefillUrl({
			title: 'Read *later tomorrow',
			url: 'https://example.com/article?a=1&b=2',
			description: 'Some <b>notes</b>',
		}))

		await expect(page).toHaveURL('/')
		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toHaveValue('"Read *later tomorrow"')
		await expect(page.locator('.quick-actions .prefill-description')).toContainText('https://example.com/article?a=1&b=2')

		await input.press('Enter')

		const task = await createdTask(page, apiContext, token)
		expect(task.project_id).toBe(project.id)
		// Quoted title: no label, no due date parsed from the page title.
		expect(task.title).toBe('Read *later tomorrow')
		expect(task.labels ?? []).toEqual([])
		expect(task.due_date).toBe('0001-01-01T00:00:00Z')
		expect(task.description).toContain('<a href="https://example.com/article?a=1&amp;b=2"')
		expect(task.description).toContain('<p>Some &lt;b&gt;notes&lt;/b&gt;</p>')
	})

	test('Creates a task without description when the prefilled description is removed', async ({page, apiContext}) => {
		const {user} = await seedUser()
		const {token} = await login(page, apiContext, user)

		await page.goto(prefillUrl({title: 'Some page', url: 'https://example.com'}))

		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toHaveValue('"Some page"')
		await page.getByRole('button', {name: 'Remove the description'}).click()
		await expect(page.locator('.quick-actions .prefill-description')).toHaveCount(0)
		await input.press('Enter')

		const task = await createdTask(page, apiContext, token)
		expect(task.title).toBe('Some page')
		expect(task.description ?? '').toBe('')
	})

	test('Asks for a title when only a url is passed', async ({page, apiContext}) => {
		const {user} = await seedUser()
		const {token} = await login(page, apiContext, user)

		await page.goto(prefillUrl({url: 'https://example.com/only-url'}))

		await expect(page).toHaveURL('/')
		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toHaveValue('')
		await expect(input).toBeFocused()
		await expect(page.locator('.quick-actions .prefill-description')).toContainText('https://example.com/only-url')

		// Enter without a title creates nothing and keeps the dialog open.
		await input.press('Enter')
		await expect(page).toHaveURL('/')
		await expect(page.locator('.quick-actions .prefill-description')).toBeVisible()
		const tasks = await apiContext.get('/api/v2/tasks', {headers: {Authorization: `Bearer ${token}`}})
		expect(tasks.ok()).toBe(true)
		expect((await tasks.json()).items).toEqual([])

		await input.fill('Read this')
		await input.press('Enter')

		const task = await createdTask(page, apiContext, token)
		expect(task.title).toBe('Read this')
		expect(task.description).toBe('<p><a href="https://example.com/only-url">https://example.com/only-url</a></p>')
	})

	test('Creates the task even when the title matches a single project', async ({page, apiContext}) => {
		const {user} = await seedUser()
		const {token} = await login(page, apiContext, user)

		await page.goto(prefillUrl({title: 'Inbox', url: 'https://example.com'}))

		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toHaveValue('"Inbox"')
		await input.press('Enter')

		const task = await createdTask(page, apiContext, token)
		expect(task.title).toBe('Inbox')
	})

	test('Keeps the prefill through login', async ({page}) => {
		const {user} = await seedUser()
		await setupApiUrl(page)

		await page.goto(prefillUrl({title: 'Shared page', url: 'https://example.com/shared'}))
		await expect(page).toHaveURL(/\/login/)

		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()

		await expect(page.locator('.quick-actions .action-input input')).toHaveValue('"Shared page"')
		await expect(page.locator('.quick-actions .prefill-description')).toContainText('https://example.com/shared')
	})
})
