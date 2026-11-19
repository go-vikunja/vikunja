import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

test('task shortcuts stay blocked while the repeat editor waits for a real save response', async ({
	authenticatedPage: page,
	apiContext,
	userToken,
}) => {
	await ProjectFactory.create(1)
	const taskId = 1
	await TaskFactory.create(1, {
		id: taskId,
		project_id: 1,
		title: 'Shortcut during repeat save',
		due_date: '2090-10-20T12:00:00Z',
		done: false,
	})
	await page.goto(`/tasks/${taskId}`)
	const trigger = page.locator('.task-view .property-repeat .repeat-after [aria-haspopup="dialog"]')
	await trigger.click()
	const editor = page.getByRole('dialog', {name: 'Repeat', exact: true})
	await expect(editor).toBeVisible()

	let release!: () => void
	const released = new Promise<void>(resolve => {
		release = resolve
	})
	let responseHeld!: () => void
	const held = new Promise<void>(resolve => {
		responseHeld = resolve
	})
	let patches = 0
	const taskPath = `/api/v2/tasks/${taskId}`
	await page.route(url => url.pathname === taskPath, async route => {
		if (route.request().method() !== 'PATCH') {
			return route.fallback()
		}
		const patchNumber = ++patches
		const response = await route.fetch()
		if (patchNumber === 1) {
			expect(response.ok(), await response.text()).toBe(true)
			responseHeld()
			await released
		}
		await route.fulfill({response})
	})
	const saved = page.waitForResponse(response =>
		new URL(response.url()).pathname === taskPath && response.request().method() === 'PATCH',
	)

	try {
		await editor.getByRole('button', {name: 'Every day', exact: true}).click()
		await held
		await expect(editor.getByRole('button', {name: 'Every day', exact: true})).toBeFocused()
		await expect(editor.getByRole('button', {name: 'Every day', exact: true})).toHaveAttribute('aria-disabled', 'true')

		for (const key of ['ControlOrMeta+s', 'Shift+T', 't', 's', 'l']) {
			await page.keyboard.press(key)
		}
		// Let an accidental shortcut request reach the route while the real response stays held.
		await page.waitForTimeout(300)
		expect(patches).toBe(1)
		await expect(editor).toBeVisible()
		await expect(page).toHaveURL(`/tasks/${taskId}`)
		await expect(page.locator('.property-labels input:focus')).toHaveCount(0)
	} finally {
		release()
	}

	expect((await saved).ok()).toBe(true)
	await expect(editor).toHaveCount(0)
	expect(patches).toBe(1)
	await page.reload()
	await expect(trigger).toContainText('Every day')
	const response = await apiContext.get(taskPath, {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(response.ok()).toBe(true)
	const stored = await response.json()
	expect(stored.repeat).toMatchObject({freq: 'daily'})
	expect(stored.done).toBe(false)
	expect(stored.is_favorite).toBe(false)
})
