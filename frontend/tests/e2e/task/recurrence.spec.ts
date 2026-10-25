import type {APIRequestContext, Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

const DAY_MS = 86_400_000

// The fixture's apiContext is rooted at /api/v1; an absolute path reaches v2 on the same origin.
async function getTaskV2(apiContext: APIRequestContext, userToken: string, id: number) {
	const response = await apiContext.get(`/api/v2/tasks/${id}`, {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(response.ok()).toBe(true)
	return response.json()
}

// Resolves with the v2 PATCH the task detail sends for the task. Register it before the action.
function waitForTaskPatch(page: Page, id: number) {
	return page.waitForResponse(r =>
		new URL(r.url()).pathname === `/api/v2/tasks/${id}` && r.request().method() === 'PATCH',
	)
}

async function expectPatchOk(patch: ReturnType<typeof waitForTaskPatch>) {
	expect((await patch).ok()).toBeTruthy()
}

// Easter Sunday (Gregorian, anonymous algorithm) of a year, as a UTC yyyy-mm-dd string.
function easterSunday(year: number): string {
	const a = year % 19
	const b = Math.floor(year / 100)
	const c = year % 100
	const d = Math.floor(b / 4)
	const e = b % 4
	const f = Math.floor((b + 8) / 25)
	const g = Math.floor((b - f + 1) / 3)
	const h = (19 * a + b - d - g + 15) % 30
	const i = Math.floor(c / 4)
	const k = c % 4
	const l = (32 + 2 * e + 2 * i - h - k) % 7
	const m = Math.floor((a + 11 * h + 22 * l) / 451)
	const month = Math.floor((h + l - 7 * m + 114) / 31)
	const day = ((h + l - 7 * m + 114) % 31) + 1
	return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

const unitSelect = (page: Page) => page.locator('.repeat-interval-unit select')
const advancedNote = (page: Page) => page.locator('.repeat-advanced-note')

test.describe('Task recurrence', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1, {id: 1})
	})

	test('sets repeat-every-day via preset button', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: new Date(Date.now() + DAY_MS).toISOString(),
		}, false)
		await page.goto(`/tasks/${task.id}`)

		// Reveal the RepeatAfter component (hidden until the user activates it)
		await page.getByRole('button', {name: 'Set Repeat'}).click()

		const save = waitForTaskPatch(page, task.id)
		await page.getByRole('button', {name: 'Every Day'}).click()
		await expectPatchOk(save)

		const stored = await getTaskV2(apiContext, userToken, task.id)
		expect(stored.repeat).toMatchObject({freq: 'daily', interval: 1})
		expect(stored.repeat.read_only).toBeFalsy()

		await page.reload()
		await expect(page.locator('#repeatInterval')).toHaveValue('1')
		await expect(unitSelect(page)).toHaveValue('days')
	})

	test('sets an interval and unit', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [task] = await TaskFactory.create(1, {id: 1, project_id: 1}, false)
		await page.goto(`/tasks/${task.id}`)
		await page.getByRole('button', {name: 'Set Repeat'}).click()

		const unitSaved = waitForTaskPatch(page, task.id)
		await unitSelect(page).selectOption('weeks')
		await expectPatchOk(unitSaved)

		const intervalSaved = waitForTaskPatch(page, task.id)
		await page.locator('#repeatInterval').fill('2')
		await page.locator('#repeatInterval').press('Tab')
		await expectPatchOk(intervalSaved)

		const stored = await getTaskV2(apiContext, userToken, task.id)
		expect(stored.repeat).toMatchObject({freq: 'weekly', interval: 2})
		expect(stored.repeat.read_only).toBeFalsy()

		await page.reload()
		await expect(page.locator('#repeatInterval')).toHaveValue('2')
		await expect(unitSelect(page)).toHaveValue('weeks')
	})

	test('completing a recurring task reopens with advanced due date', async ({
		authenticatedPage: page, apiContext, userToken,
	}) => {
		const originalDue = new Date(Date.now() + DAY_MS)
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: originalDue.toISOString(),
			repeats: 'FREQ=DAILY;INTERVAL=1',
		}, false)

		await page.goto(`/tasks/${task.id}`)

		const completed = waitForTaskPatch(page, task.id)
		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'}).click()
		await expectPatchOk(completed)

		// Fetch fresh state from the API to verify the backend regenerated the task.
		const refreshed = await getTaskV2(apiContext, userToken, task.id)
		expect(refreshed.done).toBe(false)
		expect(refreshed.repeat).toMatchObject({freq: 'daily', interval: 1})
		const newDue = new Date(refreshed.due_date).getTime()
		// A due date still in the future advances by exactly one interval (one day here).
		// Tolerance of <5s absorbs sub-second timestamp round-tripping between
		// the JS Date → ISO string → backend time.Time → JSON response path.
		expect(newDue - originalDue.getTime()).toBeCloseTo(DAY_MS, -4)
	})

	test('selecting months reveals the day select', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [task] = await TaskFactory.create(1, {id: 1, project_id: 1}, false)
		await page.goto(`/tasks/${task.id}`)
		await page.getByRole('button', {name: 'Set Repeat'}).click()

		await expect(page.locator('#repeatDay')).toHaveCount(0)

		const unitSaved = waitForTaskPatch(page, task.id)
		await unitSelect(page).selectOption('months')
		await expectPatchOk(unitSaved)
		await expect(page.locator('#repeatDay')).toBeVisible()

		const daySaved = waitForTaskPatch(page, task.id)
		await page.locator('#repeatDay').selectOption('15')
		await expectPatchOk(daySaved)

		const stored = await getTaskV2(apiContext, userToken, task.id)
		expect(stored.repeat).toMatchObject({freq: 'monthly', interval: 1, by_month_day: [15]})
		expect(stored.repeat.read_only).toBeFalsy()

		// The loaded by_month_day must come back into the day select, not just be stored.
		await page.reload()
		await expect(unitSelect(page)).toHaveValue('months')
		await expect(page.locator('#repeatDay')).toHaveValue('15')
	})

	test('removes a repeat', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			repeats: 'FREQ=WEEKLY;INTERVAL=1',
		}, false)
		await page.goto(`/tasks/${task.id}`)
		await expect(page.locator('.repeat-after-input')).toBeVisible()

		const removed = waitForTaskPatch(page, task.id)
		await page.getByRole('button', {name: 'Remove repeat interval'}).click()
		await expectPatchOk(removed)

		const stored = await getTaskV2(apiContext, userToken, task.id)
		expect(stored.repeat).toBeUndefined()

		await page.reload()
		await expect(page.locator('.task-view .heading')).toBeVisible()
		await expect(page.locator('.repeat-after-input')).toHaveCount(0)
		await expect(page.getByRole('button', {name: 'Set Repeat'})).toBeVisible()
	})

	test.describe('read-only rules', () => {
		test('shows a rule the editor cannot express as read-only', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const [task] = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
				repeats: 'FREQ=YEARLY;BYEASTER=0',
			}, false)

			const stored = await getTaskV2(apiContext, userToken, task.id)
			expect(stored.repeat).toMatchObject({freq: 'yearly', read_only: true})

			await page.goto(`/tasks/${task.id}`)
			await expect(advancedNote(page)).toBeVisible()
			await expect(advancedNote(page)).toContainText('advanced recurrence rule')
			await expect(page.getByRole('button', {name: 'Every Day'})).toBeDisabled()
			await expect(page.locator('#repeatInterval')).toBeDisabled()
			await expect(unitSelect(page)).toBeDisabled()
		})

		test('renaming the task keeps the read-only rule', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			// Anchor in the future (12:00 UTC, so a server time zone can't move the calendar day),
			// before that year's Easter: completion then steps from the due date to Easter.
			const year = new Date().getUTCFullYear() + 1
			const due = new Date(Date.UTC(year, 0, 10, 12))
			const [task] = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
				title: 'Before rename',
				due_date: due.toISOString(),
				repeats: 'FREQ=YEARLY;BYEASTER=0',
			}, false)
			const before = await getTaskV2(apiContext, userToken, task.id)
			expect(before.repeat.read_only).toBe(true)

			await page.goto(`/tasks/${task.id}`)
			const heading = page.locator('.task-view h1[contenteditable]')
			await expect(heading).toContainText('Before rename')
			const renamed = waitForTaskPatch(page, task.id)
			await heading.fill('After rename')
			await heading.press('Enter')
			await expectPatchOk(renamed)

			const after = await getTaskV2(apiContext, userToken, task.id)
			expect(after.title).toBe('After rename')
			expect(after.repeat).toEqual(before.repeat)

			// The raw rule text can't be read back: the v2 task omits it, and the test/ seed
			// endpoint only accepts PATCH and DELETE. read_only:true above only says the stored rule
			// still has *some* unrepresentable part. Completion is the stronger check of the text:
			// with BYEASTER=0 intact the due date lands on Easter Sunday; had the rename written
			// back the projection (plain FREQ=YEARLY), it would land on 10 January a year later.
			await page.reload()
			const completed = waitForTaskPatch(page, task.id)
			await page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'}).click()
			await expectPatchOk(completed)

			const completedTask = await getTaskV2(apiContext, userToken, task.id)
			expect(completedTask.done).toBe(false)
			expect(completedTask.repeat.read_only).toBe(true)
			expect(new Date(completedTask.due_date).toISOString().slice(0, 10)).toBe(easterSunday(year))
		})

		test('removes a read-only rule', async ({authenticatedPage: page, apiContext, userToken}) => {
			const [task] = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
				repeats: 'FREQ=YEARLY;BYEASTER=0',
			}, false)
			await page.goto(`/tasks/${task.id}`)
			await expect(advancedNote(page)).toBeVisible()

			const removed = waitForTaskPatch(page, task.id)
			await page.getByRole('button', {name: 'Remove repeat interval'}).click()
			await expectPatchOk(removed)

			const stored = await getTaskV2(apiContext, userToken, task.id)
			expect(stored.repeat).toBeUndefined()

			await page.reload()
			await expect(page.locator('.task-view .heading')).toBeVisible()
			await expect(page.locator('.repeat-after-input')).toHaveCount(0)
		})

		test('shows the migrated month-end rule as read-only', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			// The legacy "monthly on the 31st" migrates to the last of days 28-31 in each month.
			const [task] = await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
				repeats: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1',
			}, false)

			// The API can represent this rule, so it is not flagged; the editor can't, so it locks it.
			const stored = await getTaskV2(apiContext, userToken, task.id)
			expect(stored.repeat).toMatchObject({
				freq: 'monthly',
				interval: 1,
				by_month_day: [28, 29, 30, 31],
				by_set_pos: [-1],
			})
			expect(stored.repeat.read_only).toBeFalsy()

			await page.goto(`/tasks/${task.id}`)
			await expect(advancedNote(page)).toBeVisible()
			await expect(page.locator('#repeatInterval')).toBeDisabled()
			await expect(page.locator('#repeatDay')).toBeDisabled()
		})
	})
})
