import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {createDefaultViews} from '../project/prepareProjects'

const DOCK_TOOLBAR = 40
const HEADER = 64
// 3rem, where the search dialog opens without the offset.
const SEARCH_TOP = 48
const WINDOW = {width: 1280, height: 720}

async function behaveLikeDockApp(page: Page) {
	await page.addInitScript(() => {
		Object.defineProperty(navigator, 'standalone', {get: () => true})
	})
}

function top(page: Page, selector: string) {
	return page.locator(selector).first().evaluate(element => element.getBoundingClientRect().top)
}

test.describe('Header offset for the Safari Dock app', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1)
		await TaskFactory.create(1, {id: 1, index: 1, title: 'Offset check'})
		await page.setViewportSize(WINDOW)
	})

	test('a normal page keeps the header at the top', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/1')
		await expect(page.locator('.navbar')).toBeVisible()

		expect(await top(page, '.navbar')).toBe(0)
		expect(await page.evaluate(() => document.documentElement.style.getPropertyValue('--app-top-inset'))).toBe('')
	})

	test('a Dock app moves the header, sidebar and content below the toolbar', async ({authenticatedPage: page}) => {
		await behaveLikeDockApp(page)
		await page.goto('/tasks/1')
		await expect(page.locator('.navbar')).toBeVisible()

		expect(await top(page, '.navbar')).toBe(DOCK_TOOLBAR)
		expect(await top(page, '.menu-container')).toBeGreaterThanOrEqual(DOCK_TOOLBAR + HEADER)
		expect(await top(page, '.task-view h1.title')).toBeGreaterThanOrEqual(DOCK_TOOLBAR + HEADER)
	})

	test('search opens below the toolbar in a Dock app', async ({authenticatedPage: page}) => {
		await behaveLikeDockApp(page)
		await page.goto('/tasks/1')
		await page.locator('.navbar-end .trigger-button').first().click()

		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toBeFocused()
		expect(await top(page, '.quick-actions .action-input input')).toBeGreaterThanOrEqual(SEARCH_TOP + DOCK_TOOLBAR)
	})

	test('a dialog and its close button stay below the toolbar in a Dock app', async ({authenticatedPage: page}) => {
		await behaveLikeDockApp(page)
		await page.setViewportSize({width: 1000, height: 720})
		await page.goto('/tasks/1')
		await page.locator('.keyboard-shortcuts-button').click()

		const dialog = page.locator('.modal-content').filter({hasText: 'Keyboard Shortcuts'})
		await expect(dialog).toBeVisible()
		expect(await top(page, '.modal-content:has(.card)')).toBeGreaterThanOrEqual(DOCK_TOOLBAR)
		expect(await top(page, 'dialog[open] .close')).toBeGreaterThanOrEqual(DOCK_TOOLBAR)
	})

	test('a popup that opens upward stays below the moved header', async ({authenticatedPage: page}) => {
		await behaveLikeDockApp(page)
		await page.goto('/tasks/1')
		await page.locator('.property-due-date .datepicker button.show').click()
		await expect(page.locator('.datepicker-popup')).toBeVisible()

		expect(await top(page, '.datepicker-popup')).toBeGreaterThanOrEqual(DOCK_TOOLBAR + HEADER)
	})

	test('the skip link appears below the toolbar in a Dock app', async ({authenticatedPage: page}) => {
		await behaveLikeDockApp(page)
		await page.goto('/tasks/1')
		await expect(page.locator('.navbar')).toBeVisible()
		await page.locator('.skip-to-content').focus()

		expect(await top(page, '.skip-to-content')).toBeGreaterThanOrEqual(DOCK_TOOLBAR)
	})

	test('a menu in a task dialog stays below the toolbar in a short Dock window', async ({authenticatedPage: page}) => {
		const views = await createDefaultViews(1)
		const buckets = await BucketFactory.create(1, {project_view_id: views[3].id})
		await TaskBucketFactory.create(1, {task_id: 1, bucket_id: buckets[0].id, project_view_id: views[3].id})
		await behaveLikeDockApp(page)
		await page.setViewportSize({width: 1280, height: 360})
		await page.goto('/projects/1/4')
		await page.locator('.bucket .task').first().click()
		await expect(page.locator('dialog[open] .task-view')).toBeVisible()

		await page.locator('.task-actions-menu .toolbar-icon-button').click()
		await expect(page.locator('.task-actions-menu .dropdown-menu')).toBeVisible()
		expect(await top(page, '.task-actions-menu .dropdown-menu')).toBeGreaterThanOrEqual(DOCK_TOOLBAR)
	})
})
