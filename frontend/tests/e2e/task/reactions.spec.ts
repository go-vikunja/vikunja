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
		new URL(r.url()).pathname.endsWith(`/tasks/${task.id}/reactions`) && r.request().method() === 'PUT',
	)
	await option.click()
	expect((await created).ok()).toBeTruthy()

	const values = [SEEDED_REACTION, added!]
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}

	// Client-side round trip inside the 60s stale window: the cached task must carry both reactions.
	await page.locator('.task-view nav.subtitle a').first().click()
	await expect(page).toHaveURL(/\/projects\/1\/\d+/)
	await page.locator('.tasks .task').filter({hasText: task.title}).getByText(task.title, {exact: true}).click()
	await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}

	await page.reload()
	for (const value of values) {
		await expect(reactions.locator('.reaction-button').filter({hasText: value})).toBeVisible()
	}

	const stored = await apiContext.get(`tasks/${task.id}/reactions`, {headers})
	expect(stored.ok()).toBeTruthy()
	expect(Object.keys(await stored.json()).sort()).toEqual([...values].sort())
})
