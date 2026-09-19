import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

const SEEDED_REACTION = '🎉'

test('a reaction added in the task detail survives navigating away and back', async ({authenticatedPage: page, apiContext, userToken}) => {
	const headers = {Authorization: `Bearer ${userToken}`}
	const [project] = await ProjectFactory.create(1)
	await createDefaultViews(project.id)
	const [task] = await TaskFactory.create(1, {
		id: 1,
		project_id: project.id,
		title: 'Reacted task',
	})
	const seeded = await apiContext.put(`tasks/${task.id}/reactions`, {headers, data: {value: SEEDED_REACTION}})
	expect(seeded.ok()).toBeTruthy()

	await page.goto(`/tasks/${task.id}`)
	const reactions = page.locator('.task-view .reactions.details')
	await expect(reactions.locator('.reaction-button').filter({hasText: SEEDED_REACTION})).toBeVisible()

	await reactions.getByRole('button', {name: 'Add your reaction'}).click()
	const picker = page.locator('.emoji-picker')
	await expect(picker).toBeVisible()
	await picker.locator('input.search').fill('thumbs up')
	const option = picker.locator('#search-results [role="option"]').first()
	await expect(option).toBeVisible()
	const added = (await option.textContent())?.trim()
	expect(added).toBeTruthy()

	const created = page.waitForResponse(r =>
		new URL(r.url()).pathname.endsWith(`/tasks/${task.id}/reactions`) && r.request().method() === 'POST',
	)
	await option.click()
	expect((await created).ok()).toBeTruthy()

	const values = [SEEDED_REACTION, added!]
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}

	// Client-side round trip inside the 60s stale window: both reactions come from the cache, without refetching the task.
	let taskDetailRequests = 0
	page.on('request', request => {
		if (request.method() === 'GET' && new URL(request.url()).pathname.endsWith(`/api/v2/tasks/${task.id}`)) {
			taskDetailRequests++
		}
	})
	await page.locator('.task-view nav.subtitle a').first().click()
	await expect(page).toHaveURL(/\/projects\/1\/\d+/)
	await page.locator('.tasks .task').filter({hasText: task.title}).getByText(task.title, {exact: true}).click()
	await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}
	expect(taskDetailRequests).toBe(0)

	await page.reload()
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}

	const stored = await apiContext.get(`tasks/${task.id}/reactions`, {headers})
	expect(stored.ok()).toBeTruthy()
	expect(Object.keys(await stored.json()).sort()).toEqual([...values].sort())

	const removed = page.waitForResponse(r =>
		new URL(r.url()).pathname.endsWith(`/tasks/${task.id}/reactions/delete`) && r.request().method() === 'POST',
	)
	await reactions.locator('.reaction-button').filter({hasText: added!}).click()
	expect((await removed).ok()).toBeTruthy()

	await expect(reactions.locator('.reaction-button').filter({hasText: added!})).toHaveCount(0)
	await expect(reactions.locator('.reaction-button').filter({hasText: SEEDED_REACTION})).toBeVisible()

	const afterRemoval = await apiContext.get(`tasks/${task.id}/reactions`, {headers})
	expect(afterRemoval.ok()).toBeTruthy()
	expect(Object.keys(await afterRemoval.json())).toEqual([SEEDED_REACTION])
})
