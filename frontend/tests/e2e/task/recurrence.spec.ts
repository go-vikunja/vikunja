import type {APIRequestContext, Locator, Page, Request, Response} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {updateUserSettings} from '../../support/updateUserSettings'
import {calculateNearestHours} from '../../../src/helpers/time/calculateNearestHours'

// The user's time zone setting is the zone the browser and the API server run in, so the editor's
// recurrence zone, its preview and the scheduler all agree.
const ZONE = Intl.DateTimeFormat().resolvedOptions().timeZone
const NEW_YORK = 'America/New_York'
const DAY_MS = 86_400_000
const NO_SAVE_WINDOW_MS = 1500
const PRESET_COUNT = 7
const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']
const MIN_TAP_PX = 40
const PHONE = {width: 375, height: 812}
const LOCKED_SUMMARY = 'Repeats on a schedule set elsewhere'
const MWF_SUMMARY = /^Every week on Monday, Wednesday,? and Friday$/
const FOREIGN_YEARLY = {freq: 'yearly', by_month: [3, 9], by_day: ['1mo']}

type Repeat = Record<string, unknown>
type ApiTask = Record<string, unknown> & {
	title?: string
	done?: boolean
	repeat?: Repeat
	due_date?: string
	repeats_from_current_date?: boolean
}
type JsonPatchOp = {op: string, path: string, value?: unknown}

function previewFormat(zone: string) {
	return new Intl.DateTimeFormat('en', {weekday: 'short', month: 'short', day: 'numeric', timeZone: zone})
}

const PREVIEW_FORMAT = previewFormat(ZONE)
const CLOCK_FORMAT = new Intl.DateTimeFormat('en', {hour: '2-digit', minute: '2-digit', second: '2-digit', hourCycle: 'h23', timeZone: ZONE})
// The calendar names its days like dayjs's "LL" format in English.
const CALENDAR_DAY_FORMAT = new Intl.DateTimeFormat('en-US', {month: 'long', day: 'numeric', year: 'numeric'})

// Noon local time on the next Friday strictly after today, optionally within a day-of-month range.
function nextFriday(dayOfMonth?: [number, number]): Date {
	const date = new Date()
	date.setHours(12, 0, 0, 0)
	do {
		date.setDate(date.getDate() + 1)
	} while (date.getDay() !== 5 || (dayOfMonth && (date.getDate() < dayOfMonth[0] || date.getDate() > dayOfMonth[1])))
	return date
}

function addDays(date: Date, days: number): Date {
	const result = new Date(date)
	result.setDate(result.getDate() + days)
	return result
}

function englishOrdinal(n: number): string {
	const suffix = n % 100 >= 11 && n % 100 <= 13 ? 'th' : ({1: 'st', 2: 'nd', 3: 'rd'}[n % 10] ?? 'th')
	return `${n}${suffix}`
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

// The fixture's apiContext is rooted at /api/v1; an absolute path reaches v2 on the same origin.
async function getTaskV2(apiContext: APIRequestContext, userToken: string, id: number): Promise<ApiTask> {
	const response = await apiContext.get(`/api/v2/tasks/${id}`, {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(response.ok()).toBe(true)
	return response.json()
}

// Seeds through the API like the editor writes, and returns the server's own normalised form.
async function seedRepeat(
	apiContext: APIRequestContext,
	userToken: string,
	id: number,
	repeat: Repeat,
	extra: Record<string, unknown> = {},
): Promise<ApiTask> {
	const ops: JsonPatchOp[] = Object.entries({repeat, ...extra}).map(([field, value]) => ({op: 'add', path: `/${field}`, value}))
	const response = await apiContext.patch(`/api/v2/tasks/${id}`, {
		headers: {Authorization: `Bearer ${userToken}`, 'Content-Type': 'application/json-patch+json'},
		data: ops,
	})
	expect(response.ok(), await response.text()).toBe(true)
	return getTaskV2(apiContext, userToken, id)
}

async function setUserTimezone(apiContext: APIRequestContext, userToken: string, timezone: string) {
	await updateUserSettings(apiContext, userToken, {timezone})
	const user = await apiContext.get('user', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await user.json()).settings.timezone).toBe(timezone)
}

function isTaskPatch(request: Request, id: number): boolean {
	return new URL(request.url()).pathname === `/api/v2/tasks/${id}` && request.method() === 'PATCH'
}

// Resolves with the v2 PATCH the task detail sends for the task. Register it before the action.
function waitForTaskPatch(page: Page, id: number) {
	return page.waitForResponse(r => isTaskPatch(r.request(), id))
}

async function expectPatchOk(patch: Promise<Response>): Promise<Response> {
	const response = await patch
	expect(response.ok()).toBeTruthy()
	return response
}

// Chromium does not expose the body of a fetch(Request) to Playwright, so the page records what it PATCHes.
async function recordPatchBodies(page: Page) {
	await page.addInitScript(() => {
		const sent: {url: string, body: string}[] = []
		Object.assign(window, {__repeatTestPatches: sent})
		const originalFetch = window.fetch.bind(window)
		window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
			const request = new Request(input, init)
			if (request.method === 'PATCH') {
				sent.push({url: request.url, body: await request.clone().text()})
			}
			return originalFetch(request)
		}
	})
}

// The task detail writes a JSON Patch; this is the value it sent for one field.
async function patchedValue(page: Page, response: Response, field: string): Promise<unknown> {
	const sent = await page.evaluate(url => {
		const all = (window as unknown as {__repeatTestPatches: {url: string, body: string}[]}).__repeatTestPatches
		return all.filter(entry => entry.url === url).map(entry => entry.body)
	}, response.url())
	expect(sent.length, `PATCH bodies recorded for ${response.url()}`).toBeGreaterThan(0)
	const ops = JSON.parse(sent[sent.length - 1]!) as JsonPatchOp[]
	const op = ops.find(candidate => candidate.path === `/${field}`)
	return op === undefined ? undefined : op.value
}

function withoutReadOnly(repeat: Repeat | undefined): Repeat | undefined {
	if (repeat === undefined) {
		return undefined
	}
	const {read_only: _readOnly, ...rest} = repeat
	return rest
}

// Collects PATCHes to the given tasks from now on; stop() returns what was sent during a bounded window.
function watchTaskPatches(page: Page, ...ids: number[]) {
	const seen: string[] = []
	const listener = (request: Request) => {
		if (ids.some(id => isTaskPatch(request, id))) {
			seen.push(`${request.url()} ${request.postData() ?? ''}`)
		}
	}
	page.on('request', listener)
	return {
		async stop(): Promise<string[]> {
			await page.waitForTimeout(NO_SAVE_WINDOW_MS)
			page.off('request', listener)
			return seen
		},
	}
}

async function createTask(id: number, dueDate: Date | null, title = `Recurrence task ${id}`, extra: Record<string, unknown> = {}) {
	await TaskFactory.create(1, {
		id,
		project_id: 1,
		title,
		...(dueDate === null ? {} : {due_date: dueDate.toISOString()}),
		...extra,
	}, false)
}

function heading(page: Page): Locator {
	return page.locator('.task-view h1[contenteditable]')
}

async function openTask(page: Page, id: number) {
	await page.goto(`/tasks/${id}`)
	await expect(heading(page)).toBeVisible()
}

// Moves the app to another task inside the same document, as an in-app link does, so Back stays in the app.
async function navigateInApp(page: Page, path: string) {
	await page.evaluate(target => {
		window.history.pushState(null, '', target)
		window.dispatchEvent(new PopStateEvent('popstate', {state: null}))
	}, path)
}

function repeatField(page: Page): Locator {
	return page.locator('.repeat-after')
}

function trigger(page: Page, name?: string | RegExp): Locator {
	const opener = page.locator('[aria-haspopup="dialog"]')
	return name === undefined
		? repeatField(page).getByRole('button').and(opener)
		: repeatField(page).getByRole('button', {name, exact: typeof name === 'string' ? true : undefined}).and(opener)
}

// The presets, Custom form and locked panel: a popover on wide screens, a bottom sheet on phones.
function editor(page: Page): Locator {
	return page.getByRole('dialog', {name: 'Repeat', exact: true})
}

function bottomSheet(page: Page): Locator {
	return page.locator('dialog.bottom-sheet')
}

function control(scope: Locator | Page, role: Parameters<Page['getByRole']>[0], name: string): Locator {
	return scope.getByRole(role, {name, exact: true})
}

function setRepeatButton(page: Page): Locator {
	return control(page, 'button', 'Set Repeat')
}

function removeRepeatButton(page: Page): Locator {
	return control(page, 'button', 'Remove repeat interval')
}

function previewDates(page: Page): Locator {
	return repeatField(page).locator('.repeat-details .repeat-preview-date')
}

function foreignNote(page: Page): Locator {
	return repeatField(page).locator('.repeat-details').getByText('Set elsewhere: it can\'t be edited here.')
}

function lockedNote(scope: Locator): Locator {
	return scope.getByText('Set elsewhere: it can\'t be changed here. Remove the repeat to set a new one.')
}

function replaceDialog(page: Page): Locator {
	return page.getByRole('dialog', {name: 'Replace this repeat?'})
}

function fromCompletionCheckbox(page: Page): Locator {
	// The upstream checkbox icon adds "Checkbox" to the name.
	return repeatField(page).getByRole('checkbox', {name: /From completion date$/})
}

function markDoneButton(page: Page): Locator {
	return page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'})
}

async function expectPreview(page: Page, dates: Date[], format = PREVIEW_FORMAT) {
	await expect(previewDates(page)).toHaveCount(dates.length)
	for (const [index, date] of dates.entries()) {
		await expect(previewDates(page).nth(index)).toContainText(format.format(date))
	}
}

// An unrelated save of the whole task: the title.
async function editTitle(page: Page, id: number, title: string): Promise<Response> {
	const saved = waitForTaskPatch(page, id)
	await heading(page).fill(title)
	await heading(page).press('Enter')
	return expectPatchOk(saved)
}

function focusIsInside(page: Page, selector: string): Promise<boolean> {
	return page.evaluate(sel => document.activeElement?.closest(sel) != null, selector)
}

// WCAG contrast of the element's text against the first opaque background among it and its ancestors,
// translucent layers blended in.
async function textContrast(locator: Locator): Promise<number> {
	return locator.evaluate(element => {
		type Rgba = [number, number, number, number]
		const parse = (css: string): Rgba => {
			const [r = 0, g = 0, b = 0, a = 1] = (css.match(/[\d.]+/g) ?? []).map(Number)
			return [r, g, b, css.startsWith('rgba') ? a : 1]
		}
		const over = ([r, g, b, a]: Rgba, [br, bg, bb]: Rgba): Rgba =>
			[r * a + br * (1 - a), g * a + bg * (1 - a), b * a + bb * (1 - a), 1]
		const luminance = ([r, g, b]: Rgba) => {
			const channel = (v: number) => {
				const s = v / 255
				return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4
			}
			return 0.2126 * channel(r) + 0.7152 * channel(g) + 0.0722 * channel(b)
		}
		const layers: Rgba[] = []
		let base: Rgba | null = null
		for (let node: Element | null = element; node; node = node.parentElement) {
			const color = parse(getComputedStyle(node).backgroundColor)
			if (color[3] === 1) {
				base = color
				break
			}
			if (color[3] > 0) {
				layers.push(color)
			}
		}
		if (base === null) {
			return 0
		}
		const background = layers.reduceRight<Rgba>((below, layer) => over(layer, below), base)
		const text = over(parse(getComputedStyle(element).color), background)
		const [light, dark] = [luminance(text), luminance(background)].sort((x, y) => y - x)
		return (light! + 0.05) / (dark! + 0.05)
	})
}

function endDatePicker(scope: Locator): Locator {
	return control(scope, 'group', 'End date')
}

function calendarDay(picker: Locator, date: Date): Locator {
	return control(picker, 'button', CALENDAR_DAY_FORMAT.format(date))
}

// Pages the calendar forward until the day is shown, then picks it.
async function pickCalendarDay(picker: Locator, date: Date) {
	for (let i = 0; i < 24 && await calendarDay(picker, date).count() === 0; i++) {
		await control(picker, 'button', 'Next month').click()
	}
	await calendarDay(picker, date).click()
}

async function expectHitArea(locator: Locator, label: string) {
	const box = await locator.boundingBox()
	expect(box, `${label} has a box`).not.toBeNull()
	expect(box!.width, `${label} width`).toBeGreaterThanOrEqual(MIN_TAP_PX)
	expect(box!.height, `${label} height`).toBeGreaterThanOrEqual(MIN_TAP_PX)
}

// The activating <label> is the hit area of a native radio.
async function expectRadioHitArea(scope: Locator, name: string) {
	const radio = control(scope, 'radio', name)
	await expect(radio).toBeVisible()
	await expectHitArea(radio.locator('xpath=ancestor::label[1]'), `radio "${name}"`)
}

// A control without an accessible name fails, whichever control it is.
async function expectEveryControlNamed(scope: Locator) {
	const controls = scope.locator('button, input:not([type="hidden"]), select, textarea, [role="button"]')
	const count = await controls.count()
	expect(count).toBeGreaterThan(0)
	for (let i = 0; i < count; i++) {
		const element = controls.nth(i)
		if (!await element.isVisible()) {
			continue
		}
		await expect(element, `control ${i} of ${count}: ${await element.evaluate(e => e.outerHTML.slice(0, 120))}`)
			.toHaveAccessibleName(/\S/)
	}
}

async function expectNoHorizontalScroll(page: Page) {
	const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth)
	expect(overflow).toBeLessThanOrEqual(0)
}

async function expectInsideViewportWidth(page: Page, locator: Locator) {
	const box = await locator.boundingBox()
	const width = page.viewportSize()!.width
	expect(box).not.toBeNull()
	expect(box!.x).toBeGreaterThanOrEqual(0)
	expect(box!.x + box!.width).toBeLessThanOrEqual(width)
}

test.describe('Task recurrence', () => {
	test.beforeEach(async ({apiContext, userToken, page}) => {
		await recordPatchBodies(page)
		await ProjectFactory.create(1, {id: 1})
		await setUserTimezone(apiContext, userToken, ZONE)
	})

	test.describe('existing behaviour', () => {
		test('sets repeat-every-day via the preset', async ({authenticatedPage: page, apiContext, userToken}) => {
			await createTask(1, new Date(Date.now() + DAY_MS))
			await openTask(page, 1)

			await setRepeatButton(page).click()
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Every day').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'daily'})

			const stored = await getTaskV2(apiContext, userToken, 1)
			expect(stored.repeat).toMatchObject({freq: 'daily', interval: 1})
			expect(stored.repeat?.read_only).toBeFalsy()

			await page.reload()
			await expect(trigger(page, 'Every day')).toBeVisible()
		})

		test('sets an interval and unit in Custom', async ({authenticatedPage: page, apiContext, userToken}) => {
			await createTask(1, nextFriday())
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'combobox', 'Repeat unit').selectOption('week')
			await control(editor(page), 'spinbutton', 'Every').fill('2')
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			await expectPatchOk(save)

			const stored = await getTaskV2(apiContext, userToken, 1)
			expect(stored.repeat).toMatchObject({freq: 'weekly', interval: 2})
			expect(stored.repeat?.read_only).toBeFalsy()

			await page.reload()
			await trigger(page, /^Every 2 weeks/).click()
			await control(editor(page), 'button', 'Custom…').click()
			await expect(control(editor(page), 'spinbutton', 'Every')).toHaveValue('2')
			await expect(control(editor(page), 'combobox', 'Repeat unit')).toHaveValue('week')
		})

		test('completing a recurring task reopens with advanced due date', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			// Whole seconds, so the stored instant round-trips exactly.
			const originalDue = new Date(Math.floor((Date.now() + DAY_MS) / 1000) * 1000)
			await createTask(1, originalDue, 'Daily chore', {repeats: 'FREQ=DAILY;INTERVAL=1'})
			await openTask(page, 1)

			const completed = waitForTaskPatch(page, 1)
			await markDoneButton(page).click()
			await expectPatchOk(completed)

			// Fetch fresh state from the API to verify the backend regenerated the task.
			const refreshed = await getTaskV2(apiContext, userToken, 1)
			expect(refreshed.done).toBe(false)
			expect(refreshed.repeat).toMatchObject({freq: 'daily', interval: 1})
			const newDue = new Date(refreshed.due_date!)
			// A due date still in the future moves to the next calendar day at the same local clock time in the
			// user's zone (the runner's zone, so setDate works in it). Across a DST change that is 23 or 25 hours;
			// a clock time the change skips moves forward by the gap, as setDate also does.
			const expectedDue = addDays(originalDue, 1)
			expect(PREVIEW_FORMAT.format(newDue)).toBe(PREVIEW_FORMAT.format(expectedDue))
			expect(CLOCK_FORMAT.format(newDue)).toBe(CLOCK_FORMAT.format(expectedDue))
			// Tolerance of <5s absorbs sub-second timestamp round-tripping between
			// the JS Date → ISO string → backend time.Time → JSON response path.
			expect(Math.abs(newDue.getTime() - expectedDue.getTime())).toBeLessThan(5000)
		})

		test('choosing months in Custom reveals the day of the month', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const friday = nextFriday()
			await createTask(1, friday.getDate() === 15 ? addDays(friday, 7) : friday)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await expect(control(editor(page), 'radio', 'On day')).toHaveCount(0)
			await control(editor(page), 'combobox', 'Repeat unit').selectOption('month')
			await expect(control(editor(page), 'radio', 'On day')).toBeVisible()
			await control(editor(page), 'combobox', 'Day of the month').selectOption('15')
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			await expectPatchOk(save)

			const stored = await getTaskV2(apiContext, userToken, 1)
			expect(stored.repeat).toMatchObject({freq: 'monthly', interval: 1, by_month_day: [15]})
			expect(stored.repeat?.read_only).toBeFalsy()

			// The stored day must come back into the form, not just be stored.
			await page.reload()
			await trigger(page, 'Every month on the 15th').click()
			await control(editor(page), 'button', 'Custom…').click()
			await expect(control(editor(page), 'combobox', 'Repeat unit')).toHaveValue('month')
			await expect(control(editor(page), 'radio', 'On day')).toBeChecked()
			await expect(control(editor(page), 'combobox', 'Day of the month')).toHaveValue('15')
		})

		test('removes a repeat', async ({authenticatedPage: page, apiContext, userToken}) => {
			await createTask(1, null, 'Weekly', {repeats: 'FREQ=WEEKLY;INTERVAL=1'})
			await openTask(page, 1)
			await expect(trigger(page)).toBeVisible()

			const removed = waitForTaskPatch(page, 1)
			await removeRepeatButton(page).click()
			await expectPatchOk(removed)

			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toBeUndefined()

			await page.reload()
			await expect(heading(page)).toBeVisible()
			await expect(repeatField(page)).toHaveCount(0)
			await expect(setRepeatButton(page)).toBeVisible()
		})
	})

	test.describe('locked repeats', () => {
		test('a server read-only repeat shows a generic summary and offers no replacement', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday(), 'Easter', {repeats: 'FREQ=YEARLY;BYEASTER=0'})
			const stored = await getTaskV2(apiContext, userToken, 1)
			expect(stored.repeat).toMatchObject({freq: 'yearly', read_only: true})

			await openTask(page, 1)
			const field = trigger(page, LOCKED_SUMMARY)
			await expect(field).toBeVisible()
			await expect(lockedNote(repeatField(page).locator('.repeat-details'))).toBeVisible()
			await expect(foreignNote(page)).toHaveCount(0)

			await field.click()
			await expect(editor(page)).toBeVisible()
			await expect(editor(page)).toContainText(LOCKED_SUMMARY)
			await expect(lockedNote(editor(page))).toBeVisible()
			await expect(control(editor(page), 'button', 'Every day')).toHaveCount(0)
			await expect(control(editor(page), 'button', 'Custom…')).toHaveCount(0)
			await expect(editor(page).getByRole('button')).toHaveCount(0)

			const saves = watchTaskPatches(page, 1)
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)
			await expect(field).toBeFocused()
			expect(await saves.stop(), 'closing the locked panel saves nothing').toEqual([])
			expect(await getTaskV2(apiContext, userToken, 1)).toEqual(stored)
		})

		test('renaming the task keeps the locked repeat', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			// Anchor in the future (12:00 UTC, so a server time zone can't move the calendar day),
			// before that year's Easter: completion then steps from the due date to Easter.
			const year = new Date().getUTCFullYear() + 1
			await createTask(1, new Date(Date.UTC(year, 0, 10, 12)), 'Before rename', {repeats: 'FREQ=YEARLY;BYEASTER=0'})
			const before = await getTaskV2(apiContext, userToken, 1)
			expect(before.repeat?.read_only).toBe(true)

			await openTask(page, 1)
			await expect(heading(page)).toContainText('Before rename')
			await editTitle(page, 1, 'After rename')

			const after = await getTaskV2(apiContext, userToken, 1)
			expect(after.title).toBe('After rename')
			expect(after.repeat).toEqual(before.repeat)

			// The raw rule text can't be read back: the v2 task omits it, and the test/ seed
			// endpoint only accepts PATCH and DELETE. read_only:true above only says the stored rule
			// still has *some* unrepresentable part. Completion is the stronger check of the text:
			// with BYEASTER=0 intact the due date lands on Easter Sunday; had the rename written
			// back the projection (plain FREQ=YEARLY), it would land on 10 January a year later.
			await page.reload()
			const completed = waitForTaskPatch(page, 1)
			await markDoneButton(page).click()
			await expectPatchOk(completed)

			const completedTask = await getTaskV2(apiContext, userToken, 1)
			expect(completedTask.done).toBe(false)
			expect(completedTask.repeat?.read_only).toBe(true)
			expect(new Date(completedTask.due_date!).toISOString().slice(0, 10)).toBe(easterSunday(year))
		})

		test('toggling "From completion date" sends the locked repeat back unchanged', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday(), 'Easter', {repeats: 'FREQ=YEARLY;BYEASTER=0'})
			const before = await getTaskV2(apiContext, userToken, 1)
			expect(before.repeats_from_current_date).toBe(false)
			await openTask(page, 1)

			const save = waitForTaskPatch(page, 1)
			await repeatField(page).locator('.repeat-from-completion').getByText('From completion date').click()
			await expectPatchOk(save)

			await page.reload()
			const after = await getTaskV2(apiContext, userToken, 1)
			expect(after.repeats_from_current_date).toBe(true)
			expect(after.repeat).toEqual(before.repeat)
			await expect(fromCompletionCheckbox(page)).toBeChecked()
			await expect(trigger(page, LOCKED_SUMMARY)).toBeVisible()
		})

		test('removes a locked repeat', async ({authenticatedPage: page, apiContext, userToken}) => {
			await createTask(1, null, 'Easter', {repeats: 'FREQ=YEARLY;BYEASTER=0'})
			await openTask(page, 1)
			await expect(trigger(page, LOCKED_SUMMARY)).toBeVisible()

			const removed = waitForTaskPatch(page, 1)
			await removeRepeatButton(page).click()
			await expectPatchOk(removed)

			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toBeUndefined()

			await page.reload()
			await expect(heading(page)).toBeVisible()
			await expect(repeatField(page)).toHaveCount(0)
		})
	})

	test.describe('foreign repeats', () => {
		test('the migrated month-end repeat is foreign and is replaced only after confirmation', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			// The legacy "monthly on the 31st" migrates to the last of days 28-31 in each month.
			await createTask(1, nextFriday(), 'Month end', {repeats: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1'})

			// The API can represent this rule, so it is not read-only; the editor can't, so it is foreign.
			const stored = await getTaskV2(apiContext, userToken, 1)
			expect(stored.repeat).toMatchObject({
				freq: 'monthly',
				interval: 1,
				by_month_day: [28, 29, 30, 31],
				by_set_pos: [-1],
			})
			expect(stored.repeat?.read_only).toBeFalsy()

			await openTask(page, 1)
			await expect(foreignNote(page)).toBeVisible()
			await expect(repeatField(page).getByRole('spinbutton')).toHaveCount(0)

			await trigger(page).click()
			await control(editor(page), 'button', 'Every day').click()
			await expect(replaceDialog(page)).toBeVisible()

			const saves = watchTaskPatches(page, 1)
			await control(replaceDialog(page), 'button', 'Cancel').click()
			await expect(replaceDialog(page)).toHaveCount(0)
			expect(await saves.stop(), 'no save after Cancel').toEqual([])
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(stored.repeat))

			// Cancel returns to the presets; the same fixture now accepts the replace.
			await expect(editor(page)).toBeVisible()
			await control(editor(page), 'button', 'Every day').click()
			const save = waitForTaskPatch(page, 1)
			await control(replaceDialog(page), 'button', 'Replace').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'daily'})
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toEqual({freq: 'daily', interval: 1})

			await page.reload()
			await expect(trigger(page, 'Every day')).toBeVisible()
			await expect(foreignNote(page)).toHaveCount(0)
		})

		test('a foreign repeat is summarised; replacing it with Custom opens the form without saving', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, FOREIGN_YEARLY)
			const stored = JSON.stringify(seeded.repeat)
			await openTask(page, 1)

			await expect(foreignNote(page)).toBeVisible()
			const summary = trigger(page, /^Custom repeat: /)
			await expect(summary).toBeVisible()
			await expect(summary).toContainText('March')
			await expect(summary).toContainText('September')
			await expect(summary).toContainText('Monday')

			await summary.click()
			await control(editor(page), 'button', 'Custom…').click()
			await expect(replaceDialog(page)).toBeVisible()

			const saves = watchTaskPatches(page, 1)
			await control(replaceDialog(page), 'button', 'Replace').click()
			await expect(replaceDialog(page)).toHaveCount(0)
			await expect(control(editor(page), 'spinbutton', 'Every')).toBeVisible()
			await expect(control(editor(page), 'button', 'Done')).toBeVisible()
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)
			expect(await saves.stop(), 'Replace into Custom, then Escape, saves nothing').toEqual([])
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(stored)
		})

		test('toggling "From completion date" on a foreign repeat keeps the rule', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, FOREIGN_YEARLY)
			expect(seeded.repeats_from_current_date).toBe(false)
			await openTask(page, 1)

			const save = waitForTaskPatch(page, 1)
			await repeatField(page).locator('.repeat-from-completion').getByText('From completion date').click()
			await expectPatchOk(save)
			await page.reload()

			const after = await getTaskV2(apiContext, userToken, 1)
			expect(after.repeats_from_current_date).toBe(true)
			expect(JSON.stringify(after.repeat)).toBe(JSON.stringify(seeded.repeat))
			await expect(fromCompletionCheckbox(page)).toBeChecked()
			await expect(foreignNote(page)).toBeVisible()
		})
	})

	test.describe('user stories', () => {
		test('Mon, Wed and Fri in at most five clicks, saved and still editable after reload', async ({
			authenticatedPage: page,
		}) => {
			const due = nextFriday()
			await createTask(1, due)
			await openTask(page, 1)

			let clicks = 0
			const click = async (locator: Locator) => {
				clicks++
				await locator.click()
			}

			await click(setRepeatButton(page))
			await click(control(editor(page), 'button', 'Custom…'))
			await expect(control(editor(page), 'button', 'Friday')).toHaveAttribute('aria-pressed', 'true')
			await click(control(editor(page), 'button', 'Monday'))
			await click(control(editor(page), 'button', 'Wednesday'))
			const save = waitForTaskPatch(page, 1)
			await click(control(editor(page), 'button', 'Done'))
			const response = await expectPatchOk(save)

			expect(clicks, 'click budget from the task view').toBeLessThanOrEqual(5)
			expect(await patchedValue(page, response, 'repeat')).toStrictEqual({freq: 'weekly', by_day: ['mo', 'we', 'fr']})

			await expect(editor(page)).toHaveCount(0)
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
			await expectPreview(page, [due, addDays(due, 3), addDays(due, 5)])

			// The saved repeat comes back from the API and still classifies as editable.
			await page.reload()
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
			await expect(foreignNote(page)).toHaveCount(0)
			await expectPreview(page, [due, addDays(due, 3), addDays(due, 5)])
			await trigger(page, MWF_SUMMARY).click()
			await control(editor(page), 'button', 'Custom…').click()
			for (const day of ['Monday', 'Wednesday', 'Friday']) {
				await expect(control(editor(page), 'button', day)).toHaveAttribute('aria-pressed', 'true')
			}
			await expect(control(editor(page), 'button', 'Tuesday')).toHaveAttribute('aria-pressed', 'false')
		})

		test('presets are worded from a 4th-Friday due date; picking one saves it', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday([22, 28])
			await createTask(1, due)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			const yearLabel = new Intl.DateTimeFormat('en', {month: 'short', day: 'numeric', timeZone: 'UTC'})
				.format(Date.UTC(2024, due.getMonth(), due.getDate(), 12))
			const expected = [
				'Does not repeat',
				'Every day',
				'Every weekday (Mon–Fri)',
				'Every week on Friday',
				`Every month on the ${englishOrdinal(due.getDate())}`,
				'Every month on the 4th Friday',
				`Every year on ${yearLabel}`,
				'Custom…',
			]
			await expect(editor(page).getByRole('button')).toHaveText(expected)
			for (const label of expected) {
				await expect(control(editor(page), 'button', label)).toBeVisible()
			}

			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Every month on the 4th Friday').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'monthly', by_day: ['4fr']})

			await expect(editor(page)).toHaveCount(0)
			await expect(trigger(page, 'Every month on the 4th Friday')).toBeVisible()
			await expect(previewDates(page)).toHaveCount(3)
			for (const date of await previewDates(page).allTextContents()) {
				expect(date.trim()).toMatch(/^Fri,/)
			}

			await page.reload()
			await expect(trigger(page, 'Every month on the 4th Friday')).toBeVisible()
			await expect(foreignNote(page)).toHaveCount(0)
			await trigger(page, 'Every month on the 4th Friday').click()
			await expect(control(editor(page), 'button', 'Every month on the 4th Friday')).toHaveAttribute('aria-current', 'true')

			// "Does not repeat" removes the repeat.
			const removal = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Does not repeat').click()
			expect(await patchedValue(page, await expectPatchOk(removal), 'repeat')).toBeNull()
			await page.reload()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toBeUndefined()
		})

		// A controlled browser clock pins "now", so the exact first Monday is asserted (strictly after now, at the
		// date shortcuts' default hour). Monday 08:30 is the case where the first Monday is today at 09:00.
		const noDueDateCases = [
			{name: 'on a Monday at 08:30', weekday: 1, hour: 8, minute: 30, expectedDaysAhead: 0, expectedHour: 9},
			{name: 'on a Thursday at 14:10', weekday: 4, hour: 14, minute: 10, expectedDaysAhead: 4, expectedHour: 15},
		]
		for (const scenario of noDueDateCases) {
			test(`a repeat on a task without a due date sets the first Monday and says so (${scenario.name})`, async ({
				authenticatedPage: page, apiContext, userToken,
			}) => {
				// The most recent such weekday before today keeps the clock within the login token's lifetime.
				const fakeNow = new Date()
				fakeNow.setHours(scenario.hour, scenario.minute, 0, 0)
				do {
					fakeNow.setDate(fakeNow.getDate() - 1)
				} while (fakeNow.getDay() !== scenario.weekday)
				const expectedDue = addDays(fakeNow, scenario.expectedDaysAhead)
				expectedDue.setHours(scenario.expectedHour, 0, 0, 0)
				const today = WEEKDAYS[scenario.weekday]!

				await page.clock.install({time: fakeNow})
				await createTask(1, null)
				await openTask(page, 1)

				await setRepeatButton(page).click()
				// Presets are worded from today.
				await expect(control(editor(page), 'button', `Every week on ${today}`)).toBeVisible()
				await control(editor(page), 'button', 'Custom…').click()
				await expect(control(editor(page), 'button', today)).toHaveAttribute('aria-pressed', 'true')

				const save = waitForTaskPatch(page, 1)
				if (today === 'Monday') {
					// The implicit weekday is today's, so nothing is stored for it.
					await control(editor(page), 'button', 'Done').click()
				} else {
					await control(editor(page), 'button', today).click()
					await control(editor(page), 'button', 'Monday').click()
					await control(editor(page), 'button', 'Done').click()
				}
				const response = await expectPatchOk(save)

				expect(await patchedValue(page, response, 'repeat')).toStrictEqual(today === 'Monday' ? {freq: 'weekly'} : {freq: 'weekly', by_day: ['mo']})
				expect(new Date(await patchedValue(page, response, 'due_date') as string).toISOString()).toBe(expectedDue.toISOString())
				expect(calculateNearestHours(fakeNow)).toBe(scenario.expectedHour)
				await expect(repeatField(page).getByRole('status'))
					.toHaveText(`The due date was set to ${PREVIEW_FORMAT.format(expectedDue)} so that the repeat can start.`)
				expect(new Date((await getTaskV2(apiContext, userToken, 1)).due_date!).toISOString()).toBe(expectedDue.toISOString())
			})
		}

		test('"After 3 times" previews three dates; "After 1 time" previews one', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday()
			await createTask(1, due)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'radio', 'After').check()
			await control(editor(page), 'spinbutton', 'Number of times').fill('3')
			await expect(editor(page).locator('.repeat-custom-summary')).toHaveText(/, 3 times$/)
			let save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'weekly', count: 3})
			await expect(trigger(page, /, 3 times$/)).toBeVisible()
			await expectPreview(page, [due, addDays(due, 7), addDays(due, 14)])

			await page.reload()
			await expect(trigger(page, /, 3 times$/)).toBeVisible()
			await expect(foreignNote(page)).toHaveCount(0)
			await expectPreview(page, [due, addDays(due, 7), addDays(due, 14)])

			await trigger(page, /, 3 times$/).click()
			await control(editor(page), 'button', 'Custom…').click()
			await expect(control(editor(page), 'radio', 'After')).toBeChecked()
			await control(editor(page), 'spinbutton', 'Number of times').fill('1')
			save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'weekly', count: 1})
			await page.reload()
			await expect(trigger(page, /, 1 time$/)).toBeVisible()
			await expectPreview(page, [due])
			expect((await getTaskV2(apiContext, userToken, 1)).repeat?.count).toBe(1)
		})

		test('the end date is a date-only calendar: required, never before the due date, and inclusive', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday()
			const endDay = addDays(due, 14)
			await createTask(1, due)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'radio', 'On').check()
			const picker = endDatePicker(editor(page))
			await expect(picker).toBeVisible()
			await expect(editor(page).getByLabel('Hours')).toHaveCount(0)
			await expect(editor(page).getByLabel('Minutes')).toHaveCount(0)
			const done = control(editor(page), 'button', 'Done')
			await expect(editor(page).getByText('Choose an end date.')).toBeVisible()
			await expect(done).toBeDisabled()

			await pickCalendarDay(picker, addDays(due, -1))
			await expect(editor(page).getByText('The end date can\'t be before the due date.')).toBeVisible()
			await expect(done).toBeDisabled()

			await pickCalendarDay(picker, endDay)
			await expect(done).toBeEnabled()
			await control(editor(page), 'button', 'Clear end date').click()
			await expect(editor(page).getByText('Choose an end date.')).toBeVisible()
			await expect(done).toBeDisabled()

			await pickCalendarDay(picker, endDay)
			await expect(editor(page).locator('.repeat-custom-summary')).toHaveText(/, until /)
			const save = waitForTaskPatch(page, 1)
			await done.click()
			const sent = await patchedValue(page, await expectPatchOk(save), 'repeat') as Repeat
			expect(sent).toMatchObject({freq: 'weekly'})
			expect(typeof sent.until).toBe('string')

			await page.reload()
			await expect(trigger(page, /, until /)).toBeVisible()
			// Inclusive: the end day itself is the third date, nothing after it.
			await expectPreview(page, [due, addDays(due, 7), endDay])
			const stored = await getTaskV2(apiContext, userToken, 1)
			const until = new Date(stored.repeat!.until as string)
			expect(PREVIEW_FORMAT.format(until)).toBe(PREVIEW_FORMAT.format(endDay))
			// The last second of the picked day in the recurrence zone.
			expect(PREVIEW_FORMAT.format(new Date(until.getTime() + 1000))).toBe(PREVIEW_FORMAT.format(addDays(endDay, 1)))
		})

		test('limits the server refuses are explained in the form and never saved', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			const done = control(editor(page), 'button', 'Done')
			const every = control(editor(page), 'spinbutton', 'Every')
			await control(editor(page), 'combobox', 'Repeat unit').selectOption('year')

			await every.fill('101')
			await expect(editor(page).getByText('Enter 100 or less.')).toBeVisible()
			await expect(every).toHaveAttribute('aria-invalid', 'true')
			await expect(done).toBeDisabled()
			await every.fill('100')
			await expect(editor(page).getByText('Enter 100 or less.')).toHaveCount(0)
			await expect(done).toBeEnabled()

			const count = control(editor(page), 'spinbutton', 'Number of times')
			await control(editor(page), 'radio', 'After').check()
			await count.fill('10001')
			await expect(editor(page).getByText(/^Enter 10,?000 times or less\.$/)).toBeVisible()
			await expect(done).toBeDisabled()
			await count.fill('10000')
			await expect(editor(page).getByText(/^Enter 10,?000 times or less\.$/)).toHaveCount(0)
			await expect(done).toBeEnabled()

			await count.fill('10001')
			await expect(done).toBeDisabled()
			const saves = watchTaskPatches(page, 1)
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)
			expect(await saves.stop(), 'a refused form saves nothing').toEqual([])
			await page.reload()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toBeUndefined()
		})

		test('raising only the count is refused; lowering it and removing the limit persist', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', count: 3})
			expect(seeded.repeat).toMatchObject({freq: 'weekly', count: 3})
			await openTask(page, 1)

			await trigger(page, /, 3 times$/).click()
			await control(editor(page), 'button', 'Custom…').click()
			const count = control(editor(page), 'spinbutton', 'Number of times')
			const done = control(editor(page), 'button', 'Done')
			await expect(count).toHaveValue('3')
			await count.fill('4')
			await expect(editor(page).getByText('The number of times left can\'t be raised here. Remove the repeat and set it again.')).toBeVisible()
			await expect(done).toBeDisabled()

			await count.fill('2')
			await expect(done).toBeEnabled()
			let save = waitForTaskPatch(page, 1)
			await done.click()
			await expectPatchOk(save)
			await page.reload()
			await expect(trigger(page, /, 2 times$/)).toBeVisible()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toMatchObject({freq: 'weekly', count: 2})

			await trigger(page, /, 2 times$/).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'radio', 'Never').check()
			save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			await expectPatchOk(save)
			await page.reload()
			await expect(trigger(page, /^Every week/)).toBeVisible()
			await expect(trigger(page, /times?$/)).toHaveCount(0)
			const after = await getTaskV2(apiContext, userToken, 1)
			expect(after.repeat?.freq).toBe('weekly')
			expect(after.repeat?.count ?? 0).toBe(0)
		})
	})

	test.describe('preview across a daylight-saving change', () => {
		test.use({timezoneId: NEW_YORK})

		test('a daily 00:30 repeat in New York shows three consecutive dates across the fall-back night', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await setUserTimezone(apiContext, userToken, NEW_YORK)
			// The Saturday before the first Sunday of November, at 00:30 EDT (04:30 UTC), still ahead of now.
			let year = new Date().getUTCFullYear()
			let due: Date
			for (;;) {
				const firstSunday = 1 + (7 - new Date(Date.UTC(year, 10, 1)).getUTCDay()) % 7
				due = new Date(Date.UTC(year, 10, firstSunday - 1, 4, 30))
				if (due.getTime() > Date.now() + 3_600_000) {
					break
				}
				year++
			}
			await createTask(1, due, 'Night shift', {repeats: 'FREQ=DAILY'})
			await openTask(page, 1)

			const newYork = previewFormat(NEW_YORK)
			await expect(trigger(page, 'Every day')).toBeVisible()
			// A fixed 24-hour step would land on 23:30 on Sunday and show Sunday twice.
			await expectPreview(page, [due, new Date(due.getTime() + DAY_MS), new Date(due.getTime() + 2 * DAY_MS + 3_600_000)], newYork)
			const shown = await previewDates(page).allTextContents()
			expect(new Set(shown.map(text => text.replace(' ·', '').trim())).size).toBe(3)
		})
	})

	test.describe('drafts stay with their task', () => {
		async function seedTwoTasks(apiContext: APIRequestContext, userToken: string) {
			await createTask(1, nextFriday(), 'Task B')
			await createTask(2, nextFriday(), 'Task A')
			const b = await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', by_day: ['mo', 'we', 'fr']})
			const a = await seedRepeat(apiContext, userToken, 2, {freq: 'daily'})
			return {b, a}
		}

		test('Back with an open Custom draft discards it and leaves the other task untouched', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const {b, a} = await seedTwoTasks(apiContext, userToken)
			await openTask(page, 1)
			await navigateInApp(page, '/tasks/2')
			await expect(heading(page)).toHaveText('Task A')

			await trigger(page, 'Every day').click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'spinbutton', 'Every').fill('3')

			const saves = watchTaskPatches(page, 1, 2)
			await page.goBack()
			await expect(heading(page)).toHaveText('Task B')
			await expect(editor(page)).toHaveCount(0)
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
			expect(await saves.stop(), 'navigating away saves nothing').toEqual([])

			// A later save of the other task carries its own repeat, not the draft.
			const unrelated = await editTitle(page, 1, 'Task B renamed')
			expect(await patchedValue(page, unrelated, 'repeat')).toEqual(withoutReadOnly(b.repeat))
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(b.repeat))
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 2)).repeat)).toBe(JSON.stringify(a.repeat))

			// Positive control on the same fixture: the same change, confirmed on the task it started on, is saved there.
			await page.goForward()
			await expect(heading(page)).toHaveText('Task A')
			await trigger(page, 'Every day').click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'spinbutton', 'Every').fill('3')
			const others = watchTaskPatches(page, 1)
			const save = waitForTaskPatch(page, 2)
			await control(editor(page), 'button', 'Done').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'daily', interval: 3})
			expect(await others.stop(), 'the commit touches only its own task').toEqual([])
			expect((await getTaskV2(apiContext, userToken, 2)).repeat).toMatchObject({freq: 'daily', interval: 3})
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(b.repeat))
		})

		test('a save that answers after moving to another task never changes what that task shows', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const {b} = await seedTwoTasks(apiContext, userToken)
			await openTask(page, 1)
			await navigateInApp(page, '/tasks/2')
			await expect(heading(page)).toHaveText('Task A')

			let release!: () => void
			const released = new Promise<void>(resolve => {
				release = resolve
			})
			await page.route(url => url.pathname === '/api/v2/tasks/2', async route => {
				if (route.request().method() !== 'PATCH') {
					return route.fallback()
				}
				await released
				await route.continue()
			})

			const sent = page.waitForRequest(request => isTaskPatch(request, 2))
			await trigger(page, 'Every day').click()
			await control(editor(page), 'button', 'Every week on Friday').click()
			await sent

			await page.goBack()
			await expect(heading(page)).toHaveText('Task B')
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()

			const others = watchTaskPatches(page, 1)
			const late = waitForTaskPatch(page, 2)
			release()
			const response = await expectPatchOk(late)
			expect(await patchedValue(page, response, 'repeat')).toStrictEqual({freq: 'weekly'})
			expect(await others.stop(), 'the late answer writes nothing to the shown task').toEqual([])

			await expect(heading(page)).toHaveText('Task B')
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
			// The late save landed on the task it was started on, and only there.
			expect((await getTaskV2(apiContext, userToken, 2)).repeat).toMatchObject({freq: 'weekly', interval: 1})
			const storedB = await getTaskV2(apiContext, userToken, 1)
			expect(storedB.title).toBe('Task B')
			expect(JSON.stringify(storedB.repeat)).toBe(JSON.stringify(b.repeat))

			await page.reload()
			await expect(heading(page)).toHaveText('Task B')
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
		})
	})

	test.describe('invariants', () => {
		test('a cancelled Custom edit is never carried by a later save', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', interval: 1, by_day: ['mo', 'we', 'fr']})
			await openTask(page, 1)

			await trigger(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'spinbutton', 'Every').fill('3')
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)

			const unrelated = await editTitle(page, 1, 'Renamed after a cancelled edit')
			expect(await patchedValue(page, unrelated, 'repeat')).toEqual(withoutReadOnly(seeded.repeat))
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(seeded.repeat))
			await page.reload()
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()

			// Positive control on the same fixture: the same change, confirmed, is saved.
			await trigger(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'spinbutton', 'Every').fill('3')
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'weekly', interval: 3, by_day: ['mo', 'we', 'fr']})
			await page.reload()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat?.interval).toBe(3)
		})

		test('opening, browsing and cancelling, then an unrelated save, leave every seeded repeat byte-identical', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			test.setTimeout(420_000)
			const due = nextFriday()
			// Until values stay after the relative due date.
			const untilDay = addDays(due, 250)
			const until = new Date(new Date(untilDay).setHours(23, 59, 59, 0)).toISOString().replace('.000Z', 'Z')
			let leapYear = due.getFullYear() + 1
			while (new Date(Date.UTC(leapYear, 1, 29)).getUTCMonth() !== 1) {
				leapYear++
			}
			const untilLeapDay = `${leapYear}-02-29T23:59:59Z`
			const ymd = `${untilDay.getFullYear()}-${String(untilDay.getMonth() + 1).padStart(2, '0')}-${String(untilDay.getDate()).padStart(2, '0')}`
			const untilWithOffset = `${ymd}T23:59:59.5-04:00`
			// A monthly day-15 rule on a task due another day; a week later is never the 15th again.
			const dueNot15th = due.getDate() === 15 ? addDays(due, 7) : due
			expect(dueNot15th.getDate()).not.toBe(15)
			type Kind = 'editable' | 'foreign' | 'locked'
			type Fixture = {name: string, group: string, kind: Kind, due: Date | null, repeat?: Repeat, rule?: string}
			const fixtures: Fixture[] = [
				{name: 'hourly interval 3', group: 'hourly', kind: 'editable', repeat: {freq: 'hourly', interval: 3}, due},
				{name: 'daily', group: 'daily', kind: 'editable', repeat: {freq: 'daily', interval: 1}, due},
				{name: 'daily interval 2 count 5', group: 'daily', kind: 'editable', repeat: {freq: 'daily', interval: 2, count: 5}, due},
				{name: 'weekly implicit', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1}, due},
				{name: 'weekly mo/we/fr', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, by_day: ['mo', 'we', 'fr']}, due},
				{name: 'weekly interval 2 tu/th', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 2, by_day: ['tu', 'th']}, due},
				{name: 'monthly implicit', group: 'monthly', kind: 'editable', repeat: {freq: 'monthly', interval: 1}, due},
				{name: 'monthly on the 15th', group: 'monthly', kind: 'editable', repeat: {freq: 'monthly', interval: 1, by_month_day: [15]}, due: dueNot15th},
				{name: 'monthly last day', group: 'monthly', kind: 'editable', repeat: {freq: 'monthly', interval: 1, by_month_day: [-1]}, due},
				{name: 'monthly 4th friday', group: 'monthly', kind: 'editable', repeat: {freq: 'monthly', interval: 1, by_day: ['4fr']}, due},
				{name: 'monthly last friday', group: 'monthly', kind: 'editable', repeat: {freq: 'monthly', interval: 1, by_day: ['-1fr']}, due},
				{name: 'yearly', group: 'yearly', kind: 'editable', repeat: {freq: 'yearly', interval: 1}, due},
				{name: 'weekly until', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, until}, due},
				{name: 'until on a leap day', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, until: untilLeapDay}, due},
				{name: 'until with an offset', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, until: untilWithOffset}, due},
				{name: 'explicit interval 1 and wkst mo', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, wkst: 'mo'}, due},
				{name: 'no due date', group: 'weekly', kind: 'editable', repeat: {freq: 'weekly', interval: 1, by_day: ['mo']}, due: null},
				{name: 'upper-case codes', group: 'weekly', kind: 'editable', repeat: {freq: 'WEEKLY', interval: 1, by_day: ['FR', 'MO']}, due},
				{name: 'empty advanced lists', group: 'daily', kind: 'editable', repeat: {freq: 'daily', interval: 1, by_day: [], by_month: []}, due},
				{name: 'yearly in March and September on the 1st Monday', group: 'foreign', kind: 'foreign', repeat: {freq: 'yearly', interval: 1, by_month: [3, 9], by_day: ['1mo']}, due},
				{name: 'by_set_pos last Friday', group: 'foreign', kind: 'foreign', repeat: {freq: 'monthly', interval: 1, by_day: ['fr'], by_set_pos: [-1]}, due},
				{name: 'minutely', group: 'foreign', kind: 'foreign', repeat: {freq: 'minutely', interval: 5}, due},
				{name: 'weekly wkst su', group: 'foreign', kind: 'foreign', repeat: {freq: 'weekly', interval: 2, wkst: 'su'}, due},
				{name: 'monthly 5th friday', group: 'foreign', kind: 'foreign', repeat: {freq: 'monthly', interval: 1, by_day: ['5fr']}, due},
				{name: 'yearly by_year_day', group: 'foreign', kind: 'foreign', repeat: {freq: 'yearly', interval: 1, by_year_day: [100]}, due},
				{name: 'migrated month end', group: 'foreign', kind: 'foreign', rule: 'FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=28,29,30,31;BYSETPOS=-1', due},
				{name: 'yearly on Easter', group: 'locked', kind: 'locked', rule: 'FREQ=YEARLY;BYEASTER=0', due},
			]

			expect(fixtures.filter(fixture => fixture.kind === 'editable').length).toBeGreaterThanOrEqual(19)
			expect(fixtures.filter(fixture => fixture.kind !== 'editable').length).toBeGreaterThanOrEqual(5)

			const stored = new Map<number, string>()
			for (const [index, fixture] of fixtures.entries()) {
				const id = index + 1
				await createTask(id, fixture.due, `Matrix ${fixture.name}`, fixture.rule === undefined ? {} : {repeats: fixture.rule})
				const seeded = fixture.repeat === undefined
					? await getTaskV2(apiContext, userToken, id)
					: await seedRepeat(apiContext, userToken, id, fixture.repeat)
				if (fixture.name === 'monthly on the 15th') {
					expect(new Date(seeded.due_date!).getDate(), 'stored due day differs from by_month_day').not.toBe(15)
					expect(seeded.repeat?.by_month_day).toEqual([15])
				}
				expect(seeded.repeat, `${fixture.name} was stored`).toBeTruthy()
				expect(Boolean(seeded.repeat?.read_only), `${fixture.name} read-only flag`).toBe(fixture.kind === 'locked')
				stored.set(id, JSON.stringify(seeded.repeat))
			}

			for (const [index, fixture] of fixtures.entries()) {
				const id = index + 1
				await test.step(`no-op on ${fixture.name}`, async () => {
					await openTask(page, id)
					await expect(foreignNote(page)).toHaveCount(fixture.kind === 'foreign' ? 1 : 0)
					if (fixture.kind === 'locked') {
						await expect(trigger(page, LOCKED_SUMMARY)).toBeVisible()
					}
					await trigger(page).click()
					await expect(editor(page)).toBeVisible()
					if (fixture.kind === 'editable') {
						await control(editor(page), 'button', 'Custom…').click()
						await expect(control(editor(page), 'button', 'Done')).toBeVisible()
					}
					await page.keyboard.press('Escape')
					await expect(editor(page)).toHaveCount(0)

					await editTitle(page, id, `Matrix ${fixture.name} renamed`)
					expect(JSON.stringify((await getTaskV2(apiContext, userToken, id)).repeat), fixture.name).toBe(stored.get(id))
					await page.reload()
					await expect(trigger(page)).toBeVisible()
					expect(JSON.stringify((await getTaskV2(apiContext, userToken, id)).repeat), fixture.name).toBe(stored.get(id))
				})
			}

			// Positive controls: one deliberate edit per shape group changes the same fixtures.
			const seenGroups = new Set<string>()
			for (const [index, fixture] of fixtures.entries()) {
				if (seenGroups.has(fixture.group)) {
					continue
				}
				seenGroups.add(fixture.group)
				const id = index + 1
				await test.step(`deliberate edit on ${fixture.name}`, async () => {
					await openTask(page, id)
					const save = waitForTaskPatch(page, id)
					if (fixture.kind === 'locked') {
						await removeRepeatButton(page).click()
					} else if (fixture.kind === 'foreign') {
						await trigger(page).click()
						await control(editor(page), 'button', 'Every day').click()
						await control(replaceDialog(page), 'button', 'Replace').click()
					} else {
						await trigger(page).click()
						await control(editor(page), 'button', 'Custom…').click()
						await control(editor(page), 'spinbutton', 'Every').fill(String(Number(fixture.repeat!.interval) + 1))
						await control(editor(page), 'button', 'Done').click()
					}
					await expectPatchOk(save)
					await page.reload()
					const after = (await getTaskV2(apiContext, userToken, id)).repeat
					expect(JSON.stringify(after), fixture.name).not.toBe(stored.get(id))
					if (fixture.kind === 'locked') {
						expect(after).toBeUndefined()
					} else if (fixture.kind === 'foreign') {
						expect(after).toEqual({freq: 'daily', interval: 1})
					} else {
						expect(after?.interval).toBe(Number(fixture.repeat!.interval) + 1)
					}
				})
			}
			expect([...seenGroups].sort()).toEqual(['daily', 'foreign', 'hourly', 'locked', 'monthly', 'weekly', 'yearly'])
		})

		test('the remove control clears the repeat and the completion setting', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', by_day: ['mo', 'we', 'fr']}, {repeats_from_current_date: true})
			expect(seeded.repeat?.freq).toBe('weekly')
			expect(seeded.repeats_from_current_date).toBe(true)
			await openTask(page, 1)

			const save = waitForTaskPatch(page, 1)
			await removeRepeatButton(page).click()
			await expectPatchOk(save)
			await page.reload()

			const after = await getTaskV2(apiContext, userToken, 1)
			expect(after.repeat).toBeUndefined()
			expect(after.repeats_from_current_date).toBe(false)
			await expect(repeatField(page)).toHaveCount(0)
		})
	})

	test.describe('keyboard and accessible names', () => {
		test('keyboard opens, walks and closes the presets and Custom', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', by_day: ['mo', 'we', 'fr']})
			await openTask(page, 1)
			const field = trigger(page, MWF_SUMMARY)

			await removeRepeatButton(page).focus()
			await page.keyboard.press('Tab')
			await expect(field).toBeFocused()
			await expect(field).toHaveAttribute('aria-expanded', 'false')

			await page.keyboard.press('Enter')
			await expect(field).toHaveAttribute('aria-expanded', 'true')
			await expect(control(editor(page), 'button', 'Does not repeat')).toBeFocused()
			await page.keyboard.press('Tab')
			await expect(control(editor(page), 'button', 'Every day')).toBeFocused()
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)
			await expect(field).toBeFocused()
			await expect(field).toHaveAttribute('aria-expanded', 'false')

			// A closed popover is not reachable by keyboard.
			await page.keyboard.press('Tab')
			expect(await focusIsInside(page, '.repeat-popup')).toBe(false)
			await page.keyboard.press('Shift+Tab')
			await expect(field).toBeFocused()

			await page.keyboard.press('Enter')
			await expect(control(editor(page), 'button', 'Does not repeat')).toBeFocused()
			for (let i = 0; i < PRESET_COUNT; i++) {
				await page.keyboard.press('Tab')
			}
			await expect(control(editor(page), 'button', 'Custom…')).toBeFocused()
			await page.keyboard.press('Enter')
			await expect(control(editor(page), 'spinbutton', 'Every')).toBeFocused()
			await page.keyboard.press('Tab')
			await expect(control(editor(page), 'combobox', 'Repeat unit')).toBeFocused()
			// Week starts on Sunday (the default setting), so the toggles run Sunday → Saturday.
			for (const day of WEEKDAYS) {
				await page.keyboard.press('Tab')
				await expect(control(editor(page), 'button', day)).toBeFocused()
			}
			await page.keyboard.press('Space')
			await expect(control(editor(page), 'button', 'Saturday')).toHaveAttribute('aria-pressed', 'true')
			const done = control(editor(page), 'button', 'Done')
			for (let i = 0; i < 10; i++) {
				await page.keyboard.press('Tab')
				await expect(page.locator(':focus')).toHaveAccessibleName(/\S/)
				expect(await focusIsInside(page, '.repeat-popup')).toBe(true)
				if (await done.evaluate(el => el === document.activeElement)) {
					break
				}
			}
			await expect(done).toBeFocused()

			const saves = watchTaskPatches(page, 1)
			await page.keyboard.press('Escape')
			await expect(editor(page)).toHaveCount(0)
			await expect(field).toBeFocused()
			expect(await saves.stop(), 'Escape saves nothing').toEqual([])
		})

		test('the replace confirmation holds focus and Escape returns into the presets', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, FOREIGN_YEARLY)
			await openTask(page, 1)

			await removeRepeatButton(page).focus()
			await page.keyboard.press('Tab')
			await expect(trigger(page, /^Custom repeat: /)).toBeFocused()
			await page.keyboard.press('Enter')
			await expect(control(editor(page), 'button', 'Does not repeat')).toBeFocused()
			await page.keyboard.press('Tab')
			await page.keyboard.press('Enter')

			await expect(replaceDialog(page)).toBeVisible()
			const insideDialog = () => focusIsInside(page, 'dialog[open]')
			// A native modal dialog makes the page behind it inert: Tab cycles through the dialog and may pass through
			// the browser's own UI (activeElement = body), but never reaches a control of the page behind.
			const neverBehind = () => page.evaluate(() => {
				const active = document.activeElement
				return active === null || active === document.body || active.closest('dialog[open]') !== null
			})
			expect(await insideDialog()).toBe(true)
			let returnedInside = 0
			for (let i = 0; i < 8; i++) {
				await page.keyboard.press('Tab')
				expect(await neverBehind(), `focus after Tab ${i + 1}`).toBe(true)
				if (await insideDialog()) {
					returnedInside++
				}
			}
			expect(returnedInside).toBeGreaterThanOrEqual(3)
			await control(replaceDialog(page), 'button', 'Cancel').focus()

			const saves = watchTaskPatches(page, 1)
			await page.keyboard.press('Escape')
			await expect(replaceDialog(page)).toHaveCount(0)
			await expect(editor(page)).toBeVisible()
			await expect(control(editor(page), 'button', 'Every day')).toBeFocused()
			expect(await saves.stop(), 'Escape on the confirmation saves nothing').toEqual([])
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(seeded.repeat))
		})

		test('every control in the field, popover, Custom form and confirmation has a role and a name', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday()
			await createTask(1, due)
			await createTask(2, nextFriday())
			await seedRepeat(apiContext, userToken, 2, FOREIGN_YEARLY)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await expect(trigger(page, 'Does not repeat')).toBeVisible()
			await expect(trigger(page, 'Does not repeat')).toHaveAttribute('aria-haspopup', 'dialog')
			await expect(removeRepeatButton(page)).toBeVisible()
			await expect(fromCompletionCheckbox(page)).toBeVisible()
			await expectEveryControlNamed(editor(page))

			await control(editor(page), 'button', 'Custom…').click()
			const custom = editor(page)
			for (const [role, name] of [
				['spinbutton', 'Every'], ['combobox', 'Repeat unit'], ['radio', 'Never'], ['radio', 'After'], ['radio', 'On'],
				['spinbutton', 'Number of times'], ['button', 'Back'], ['button', 'Done'],
				...WEEKDAYS.map(day => ['button', day]),
			] as const) {
				await expect(control(custom, role, name)).toBeVisible()
			}
			await expectEveryControlNamed(custom)

			await control(custom, 'combobox', 'Repeat unit').selectOption('month')
			for (const [role, name] of [
				['radio', 'On day'], ['combobox', 'Day of the month'], ['radio', 'On the'], ['combobox', 'Week of the month'],
				['combobox', 'Day of the week'], ['radio', 'On the last day'],
			] as const) {
				await expect(control(custom, role, name)).toBeVisible()
			}
			await control(custom, 'radio', 'On').check()
			const picker = endDatePicker(custom)
			for (const [role, name] of [
				['button', 'Previous month'], ['button', 'Next month'], ['combobox', 'Month'], ['textbox', 'Year'],
			] as const) {
				await expect(control(picker, role, name)).toBeVisible()
			}
			await pickCalendarDay(picker, addDays(due, 14))
			await expect(control(custom, 'button', 'Clear end date')).toBeVisible()
			await expectEveryControlNamed(custom)
			await page.keyboard.press('Escape')

			await openTask(page, 2)
			await trigger(page, /^Custom repeat: /).click()
			await control(editor(page), 'button', 'Every day').click()
			await expect(replaceDialog(page)).toBeVisible()
			for (const name of ['Cancel', 'Replace', 'Close dialog']) {
				await expect(control(replaceDialog(page), 'button', name)).toBeVisible()
			}
			await expectEveryControlNamed(replaceDialog(page))
		})
	})

	test.describe('contrast', () => {
		test('a Custom form error reads at 4.5:1 or more', async ({authenticatedPage: page}) => {
			await createTask(1, nextFriday())
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'radio', 'On').check()
			const error = editor(page).getByText('Choose an end date.')
			await expect(error).toBeVisible()
			expect(await textContrast(error)).toBeGreaterThanOrEqual(4.5)
		})
	})

	test.describe('week start', () => {
		test('the "On the" weekday list starts on the user\'s week start', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const dayOfWeek = () => control(editor(page), 'combobox', 'Day of the week')
			const openMonthlyOnThe = async () => {
				await setRepeatButton(page).click()
				await control(editor(page), 'button', 'Custom…').click()
				await control(editor(page), 'combobox', 'Repeat unit').selectOption('month')
				await control(editor(page), 'radio', 'On the').check()
			}

			await updateUserSettings(apiContext, userToken, {week_start: 1})
			await openTask(page, 1)
			await openMonthlyOnThe()
			await expect(dayOfWeek().locator('option')).toHaveText(['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'])
			await page.keyboard.press('Escape')

			await updateUserSettings(apiContext, userToken, {week_start: 0})
			await openTask(page, 1)
			await openMonthlyOnThe()
			await expect(dayOfWeek().locator('option')).toHaveText(WEEKDAYS)

			// Only the order changes: the picked day is still stored by its own code.
			await dayOfWeek().selectOption({label: 'Sunday'})
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			const saved = await patchedValue(page, await expectPatchOk(save), 'repeat') as Repeat
			expect(saved).toMatchObject({freq: 'monthly'})
			expect(saved.by_day).toEqual([expect.stringMatching(/^-?\d+su$/)])
			await page.reload()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat?.by_day).toEqual(saved.by_day)
		})
	})

	test.describe('on a phone', () => {
		test.use({viewport: PHONE})

		test('the presets open in a bottom sheet and a Custom repeat saves from it', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday()
			await createTask(1, due)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			await expect(bottomSheet(page)).toBeVisible()
			await expect(editor(page)).toBeVisible()
			await expect(editor(page)).toHaveClass(/bottom-sheet/)
			await expect(editor(page).getByRole('button')).toHaveCount(PRESET_COUNT + 1)
			await expectNoHorizontalScroll(page)

			await control(editor(page), 'button', 'Custom…').click()
			await expect(control(editor(page), 'button', 'Friday')).toHaveAttribute('aria-pressed', 'true')
			await control(editor(page), 'button', 'Monday').click()
			await control(editor(page), 'button', 'Wednesday').click()
			await expectNoHorizontalScroll(page)
			const save = waitForTaskPatch(page, 1)
			await control(editor(page), 'button', 'Done').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'weekly', by_day: ['mo', 'we', 'fr']})
			await expect(bottomSheet(page)).toHaveCount(0)

			await page.reload()
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toMatchObject({freq: 'weekly', by_day: ['mo', 'we', 'fr']})
		})

		test('closing the sheet without choosing saves nothing and returns focus to the field', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, {freq: 'weekly', by_day: ['mo', 'we', 'fr']})
			await openTask(page, 1)
			const field = trigger(page, MWF_SUMMARY)

			const saves = watchTaskPatches(page, 1)
			await field.click()
			await expect(bottomSheet(page)).toBeVisible()
			await page.keyboard.press('Escape')
			await expect(bottomSheet(page)).toHaveCount(0)
			await expect(field).toBeFocused()

			await field.click()
			await control(editor(page), 'button', 'Custom…').click()
			await control(editor(page), 'spinbutton', 'Every').fill('4')
			await page.keyboard.press('Escape')
			await expect(bottomSheet(page)).toHaveCount(0)
			expect(await saves.stop(), 'closing the sheet saves nothing').toEqual([])
			await page.reload()
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(seeded.repeat))
			await expect(trigger(page, MWF_SUMMARY)).toBeVisible()
		})

		test('the replace confirmation stacks on the sheet; Escape closes it first, then the sheet', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			const seeded = await seedRepeat(apiContext, userToken, 1, FOREIGN_YEARLY)
			await openTask(page, 1)

			await trigger(page, /^Custom repeat: /).click()
			await expect(bottomSheet(page)).toBeVisible()
			await control(editor(page), 'button', 'Every day').click()
			await expect(replaceDialog(page)).toBeVisible()
			await expect(bottomSheet(page)).toBeVisible()

			const saves = watchTaskPatches(page, 1)
			await page.keyboard.press('Escape')
			await expect(replaceDialog(page)).toHaveCount(0)
			await expect(bottomSheet(page)).toBeVisible()
			await expect(control(editor(page), 'button', 'Every day')).toBeFocused()
			await page.keyboard.press('Escape')
			await expect(bottomSheet(page)).toHaveCount(0)
			expect(await saves.stop(), 'Escape twice saves nothing').toEqual([])
			expect(JSON.stringify((await getTaskV2(apiContext, userToken, 1)).repeat)).toBe(JSON.stringify(seeded.repeat))

			// Positive control: confirming over the sheet replaces the repeat and closes the sheet.
			await trigger(page, /^Custom repeat: /).click()
			await control(editor(page), 'button', 'Every day').click()
			const save = waitForTaskPatch(page, 1)
			await control(replaceDialog(page), 'button', 'Replace').click()
			expect(await patchedValue(page, await expectPatchOk(save), 'repeat')).toStrictEqual({freq: 'daily'})
			await expect(bottomSheet(page)).toHaveCount(0)
			await page.reload()
			expect((await getTaskV2(apiContext, userToken, 1)).repeat).toEqual({freq: 'daily', interval: 1})
		})

		test('the sheet is hidden beneath the replace confirmation, which reads like any phone confirmation', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			await seedRepeat(apiContext, userToken, 1, FOREIGN_YEARLY)
			await openTask(page, 1)

			await trigger(page, /^Custom repeat: /).click()
			await expect(bottomSheet(page)).toHaveCSS('opacity', '1')
			await control(editor(page), 'button', 'Every day').click()
			const dialog = replaceDialog(page)
			await expect(dialog).toBeVisible()
			await expect(bottomSheet(page)).toHaveAttribute('open')
			await expect(bottomSheet(page)).toHaveCSS('opacity', '0')

			// A guard on the confirmation's own colours. Cancel is left out: it is the stock Modal's red Cancel, as on every
			// other confirmation.
			const texts: [string, Locator][] = [
				['title', dialog.getByText('Replace this repeat?')],
				['question', dialog.getByText(/^This task repeats /)],
				['Replace', control(dialog, 'button', 'Replace')],
			]
			for (const [name, text] of texts) {
				expect(await textContrast(text), `${name} contrast`).toBeGreaterThanOrEqual(4.5)
			}

			await page.keyboard.press('Escape')
			await expect(dialog).toHaveCount(0)
			await expect(bottomSheet(page)).toHaveCSS('opacity', '1')
			await expect(control(editor(page), 'button', 'Every day')).toBeFocused()
		})

		test('opening Custom focuses the form, not a text field, so no keyboard pops up', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await createTask(1, nextFriday())
			await createTask(2, nextFriday())
			await seedRepeat(apiContext, userToken, 2, FOREIGN_YEARLY)
			const interval = () => control(editor(page), 'spinbutton', 'Every')
			// Records every element that takes focus: even a momentary focus on a text field opens the keyboard.
			const startFocusLog = () => page.evaluate(() => {
				const log: string[] = []
				Object.assign(window, {repeatFocusLog: log})
				document.addEventListener('focusin', event => log.push((event.target as Element).tagName), {capture: true})
			})
			const expectFormFocusedWithoutTextField = async (path: string) => {
				await expect(interval()).toBeVisible()
				const focused = await page.evaluate(() => (window as unknown as {repeatFocusLog: string[]}).repeatFocusLog)
				expect(focused, `${path}: nothing that opens a keyboard took focus`).not.toContain('INPUT')
				await expect(interval(), path).not.toBeFocused()
				expect(await page.evaluate(() => document.activeElement?.tagName), path).not.toMatch(/^(INPUT|TEXTAREA|SELECT)$/)
				expect(await focusIsInside(page, '.repeat-custom-panel'), `${path}: focus is inside the Custom form`).toBe(true)
				// Positive control: the interval is the first stop from there.
				await page.keyboard.press('Tab')
				await expect(interval(), path).toBeFocused()
			}

			await openTask(page, 1)
			await setRepeatButton(page).click()
			await startFocusLog()
			await control(editor(page), 'button', 'Custom…').click()
			await expectFormFocusedWithoutTextField('Custom…')

			await openTask(page, 2)
			await trigger(page, /^Custom repeat: /).click()
			await control(editor(page), 'button', 'Custom…').click()
			await startFocusLog()
			await control(replaceDialog(page), 'button', 'Replace').click()
			await expectFormFocusedWithoutTextField('Replace → Custom')
		})

		test('the sheet fits, nothing scrolls sideways and every control is at least 40 × 40', async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			const due = nextFriday()
			await createTask(1, due)
			await createTask(2, nextFriday())
			await seedRepeat(apiContext, userToken, 2, FOREIGN_YEARLY)
			await openTask(page, 1)

			await setRepeatButton(page).click()
			const sheet = editor(page)
			await expect(sheet).toBeVisible()
			await expectInsideViewportWidth(page, sheet.locator('.repeat-options-popup'))
			await expectNoHorizontalScroll(page)

			const presetButtons = sheet.getByRole('button')
			await expect(presetButtons).toHaveCount(PRESET_COUNT + 1)
			for (let i = 0; i <= PRESET_COUNT; i++) {
				await expectHitArea(presetButtons.nth(i), `preset row ${i}`)
			}

			await control(sheet, 'button', 'Custom…').click()
			await expectInsideViewportWidth(page, sheet.locator('.repeat-options-popup'))
			await expectNoHorizontalScroll(page)
			await expectHitArea(control(sheet, 'spinbutton', 'Every'), 'interval')
			await expectHitArea(control(sheet, 'combobox', 'Repeat unit'), 'unit')
			for (const day of WEEKDAYS) {
				await expectHitArea(control(sheet, 'button', day), day)
			}
			for (const name of ['Never', 'After', 'On']) {
				await expectRadioHitArea(sheet, name)
			}
			await expectHitArea(control(sheet, 'spinbutton', 'Number of times'), 'count')
			await expectHitArea(control(sheet, 'button', 'Back'), 'Back')
			await expectHitArea(control(sheet, 'button', 'Done'), 'Done')

			await control(sheet, 'combobox', 'Repeat unit').selectOption('month')
			for (const name of ['On day', 'On the', 'On the last day']) {
				await expectRadioHitArea(sheet, name)
			}
			for (const name of ['Day of the month', 'Week of the month', 'Day of the week']) {
				await expectHitArea(control(sheet, 'combobox', name), name)
			}
			await expectInsideViewportWidth(page, sheet.locator('.repeat-options-popup'))
			await expectNoHorizontalScroll(page)

			// The end-date calendar is reached by scrolling the sheet, not the page sideways; all its controls are touch-sized.
			await control(sheet, 'radio', 'On').check()
			const picker = endDatePicker(sheet)
			await picker.scrollIntoViewIfNeeded()
			await expect(picker).toBeInViewport()
			await expectHitArea(control(picker, 'button', 'Previous month'), 'previous month')
			await expectHitArea(control(picker, 'button', 'Next month'), 'next month')
			await expectHitArea(control(picker, 'combobox', 'Month'), 'month select')
			await expectHitArea(control(picker, 'textbox', 'Year'), 'year input')
			const days = picker.locator('button[data-date]')
			const dayCount = await days.count()
			expect(dayCount).toBeGreaterThanOrEqual(28)
			for (let i = 0; i < dayCount; i++) {
				await expectHitArea(days.nth(i), `calendar day ${i}`)
			}
			await pickCalendarDay(picker, addDays(due, 14))
			await expectHitArea(control(sheet, 'button', 'Clear end date'), 'clear end date')
			await expectInsideViewportWidth(page, sheet.locator('.repeat-options-popup'))
			await expectNoHorizontalScroll(page)
			await page.keyboard.press('Escape')
			await expect(bottomSheet(page)).toHaveCount(0)

			await openTask(page, 2)
			await expectHitArea(trigger(page, /^Custom repeat: /), 'field trigger')
			await trigger(page, /^Custom repeat: /).click()
			await control(editor(page), 'button', 'Every day').click()
			await expect(replaceDialog(page)).toBeVisible()
			await expectHitArea(control(replaceDialog(page), 'button', 'Cancel'), 'Cancel')
			await expectHitArea(control(replaceDialog(page), 'button', 'Replace'), 'Replace')
			await expectNoHorizontalScroll(page)
		})
	})
})
