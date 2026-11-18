import type {Locator, Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {updateUserSettings} from '../../support/updateUserSettings'

const WINDOWS = [
	{
		width: 1280,
		height: 720,
	},
	{
		width: 1440,
		height: 900,
	},
]
const DATE_ROWS = ['due-date', 'start-date', 'end-date']
const DATE = '2030-06-15T12:00:00Z'

async function expectToolbarClear(toolbar: Locator) {
	await expect(toolbar).toBeVisible()
	const {visible, covered} = await toolbar.evaluate(element => {
		const buttons = [...element.querySelectorAll('button')]
			.filter(button => getComputedStyle(button).visibility === 'visible' && button.offsetParent !== null)
		const covered = buttons.filter(button => {
			const box = button.getBoundingClientRect()
			const top = document.elementFromPoint(box.left + box.width / 2, box.top + box.height / 2)
			return top !== null && top.closest('.datepicker-popup') !== null
		})
		return {
			visible: buttons.length,
			covered: covered.length,
		}
	})

	// A missing toolbar must not make the overlap assertion pass vacuously.
	expect(visible).toBeGreaterThan(0)
	expect(covered).toBe(0)
}

async function openDescriptionEditor(page: Page) {
	await page.locator('.tiptap__task-description').getByRole('button', {
		name: 'Edit',
		exact: true,
	}).click()
	const toolbar = page.locator('.tiptap__task-description .editor-toolbar')
	await expect(toolbar).toBeVisible()
	return toolbar
}

test.describe('Date picker over the task description editor', () => {
	test.beforeEach(async ({authenticatedPage}) => {
		await ProjectFactory.create(1)
		await TaskFactory.create(1, {
			id: 1,
			index: 1,
			description: 'Some description',
			due_date: DATE,
			start_date: DATE,
			end_date: DATE,
		})
		await authenticatedPage.goto('/tasks/1')
	})

	for (const size of WINDOWS) {
		for (const row of DATE_ROWS) {
			test(`${row} leaves the formatting bar clear at ${size.width}x${size.height}`, async ({authenticatedPage: page}) => {
				await page.setViewportSize(size)
				const toolbar = await openDescriptionEditor(page)

				await page.locator(`.property-${row} .datepicker button.show`).click()
				await expect(page.locator('.datepicker-popup')).toBeVisible()

				await expectToolbarClear(toolbar)
			})
		}

		test(`a reminder date leaves the formatting bar clear at ${size.width}x${size.height}`, async ({authenticatedPage: page}) => {
			await page.setViewportSize(size)
			const toolbar = await openDescriptionEditor(page)

			await page.locator('.property-reminders').getByRole('button', {name: 'Add a reminder…'}).click()
			await page.getByRole('button', {name: 'Date and time'}).click()
			await expect(page.locator('.datepicker-popup')).toBeVisible()

			await expectToolbarClear(toolbar)
		})
	}
})

test.describe('Other date picker layouts stay unchanged', () => {
	test('settings custom token expiry keeps its non-compact calendar', async ({authenticatedPage: page}) => {
		await page.setViewportSize({
			width: 1280,
			height: 720,
		})
		await page.goto('/user/settings/api-tokens?title=Calendar')
		await page.locator('#apiTokenExpiry').selectOption('custom')
		await page.locator('.datepicker button.show').click()

		const picker = page.locator('.datepicker-popup')
		await expect(picker).toBeVisible()
		await expect(picker).toHaveClass(/datepicker-popup--no-shortcuts/)
		await expect(picker).not.toHaveClass(/datepicker-popup--compact/)
		await expect.poll(() => picker.evaluate(el => el.getBoundingClientRect().width)).toBe(380)
		await expect(picker.locator('.date-shortcuts')).toHaveCount(0)
	})

	test('filter dates keep the full range picker and sidebar', async ({authenticatedPage: page, apiContext, userToken}) => {
		await updateUserSettings(apiContext, userToken, {
			frontend_settings: {filter_editor_mode: 'advanced'},
		})
		await page.setViewportSize({
			width: 1280,
			height: 720,
		})
		await page.goto('/filters/new')
		const input = page.locator('.filter-input .ProseMirror')
		await input.fill('')
		await input.pressSequentially('dueDate = 2030-06-15', {delay: 20})
		await page.locator('.filter-input .date-value').click()

		const picker = page.locator('.filter-datepicker .datepicker-with-range')
		await expect(picker).toBeVisible()
		await expect(picker.locator('.selections')).toBeVisible()
		await expect.poll(() => picker.locator('..').evaluate(el => el.getBoundingClientRect().width)).toBe(560)
		await expect(page.locator('.datepicker-popup--compact')).toHaveCount(0)
	})

	test('task dates keep the full-width phone bottom sheet and scrolling chips', async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1)
		await TaskFactory.create(1, {
			id: 1,
			due_date: DATE,
		})
		await page.setViewportSize({
			width: 390,
			height: 844,
		})
		await page.goto('/tasks/1')
		await page.locator('.property-due-date .datepicker button.show').click()

		const sheet = page.locator('.bottom-sheet')
		const picker = sheet.locator('.datepicker-popup')
		await expect(picker).toBeVisible()
		await expect(page.locator('.popup.is-open')).toHaveCount(0)
		await expect(picker.locator('.datepicker-inline--chips')).toBeVisible()
		await expect.poll(() => picker.evaluate(el => Math.round(el.getBoundingClientRect().width))).toBe(390)
		const chips = picker.locator('.date-shortcuts--chips')
		await expect.poll(() => chips.evaluate(el => getComputedStyle(el).flexWrap)).toBe('nowrap')
		await expect.poll(() => chips.evaluate(el => getComputedStyle(el).overflowX)).toBe('auto')
		await picker.locator('[data-cy="closeDatepicker"]').click()
		await expect(sheet).not.toBeVisible()
	})
})
