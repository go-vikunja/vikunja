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
	await picker.getByRole('searchbox', {name: 'Search emoji'}).fill('thumbs up')
	const option = picker.getByRole('button', {name: 'thumbs up', exact: true})
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

test('the native picker supports categories, skin tones, recents and keyboard dismissal', async ({authenticatedPage: page, apiContext, userToken}) => {
	const [project] = await ProjectFactory.create(1)
	const [task] = await TaskFactory.create(1, {project_id: project.id})
	await page.goto(`/tasks/${task.id}`)
	const trigger = page.locator('.task-view .reactions.details').getByRole('button', {name: 'Add your reaction'})
	await trigger.click()
	const picker = page.getByRole('region', {name: 'Emoji picker'})
	const search = picker.getByRole('searchbox', {name: 'Search emoji'})
	await expect(search).toBeFocused()
	await picker.getByRole('button', {
		name: 'Animals & nature',
		exact: true,
	}).click()
	await expect(picker.getByRole('heading', {name: 'Animals & nature'})).toBeVisible()
	await picker.getByRole('combobox', {name: 'Skin tone'}).selectOption('3')
	await search.fill(':wave:')
	const wave = picker.getByRole('button', {
		name: 'waving hand',
		exact: true,
	})
	await expect(wave).toHaveText('👋🏽')
	await search.press('ArrowDown')
	await expect(wave).toBeFocused()
	await wave.press('Enter')
	await expect(picker).toHaveCount(0)
	await expect(trigger).toBeFocused()
	await page.reload()
	await trigger.click()
	await expect(picker.getByRole('combobox', {name: 'Skin tone'})).toHaveValue('3')
	await expect(picker.getByRole('heading', {name: 'Recently used'})).toBeVisible()
	await expect(picker.getByRole('button', {
		name: 'waving hand',
		exact: true,
	})).toHaveText('👋🏽')
	// Only the emojibase group 2 components (🏻, 🏼) match this query
	await search.fill('light skin tone')
	await expect(picker.getByText('No emoji found')).toBeVisible()
	await search.press('Escape')
	await expect(picker).toHaveCount(0)
	await expect(trigger).toBeFocused()
	const stored = await apiContext.get(`tasks/${task.id}/reactions`, {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(stored.ok()).toBeTruthy()
	expect(Object.keys(await stored.json())).toEqual(['👋🏽'])
})

test('the emoji picker retries a failed catalog request and dismisses on outside click', async ({authenticatedPage: page}) => {
	const [project] = await ProjectFactory.create(1)
	const [task] = await TaskFactory.create(1, {project_id: project.id})
	await page.route('**/data*.json', route => route.fulfill({
		status: 503,
		body: 'Unavailable',
	}))
	await page.goto(`/tasks/${task.id}`)
	const trigger = page.locator('.task-view .reactions.details').getByRole('button', {name: 'Add your reaction'})
	await trigger.click()
	const picker = page.getByRole('region', {name: 'Emoji picker'})
	await expect(picker.getByText('Could not load emoji.')).toBeVisible()
	await page.unroute('**/data*.json')
	await picker.getByRole('button', {name: 'Try again'}).click()
	await expect(picker.getByRole('button', {
		name: 'grinning face',
		exact: true,
	})).toBeVisible()
	await page.locator('.task-view h1').click()
	await expect(picker).toHaveCount(0)
	await expect(trigger).toHaveAttribute('aria-expanded', 'false')
	await trigger.click()
	await expect(picker).toBeVisible()
	await trigger.click()
	await expect(picker).toHaveCount(0)
})
