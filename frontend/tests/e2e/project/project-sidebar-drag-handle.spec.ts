import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {createDefaultViews} from './prepareProjects'

// Regression test for #3449: on a desktop with a touchscreen the drag handle used to be
// forced visible on top of the project color, because visibility was keyed off touch-event
// support instead of the pointer type.
test.describe('Project sidebar drag handle', () => {
	test('does not overlap the project color on a touch capable desktop', async ({authenticatedPage: page, currentUser}) => {
		const project = (await ProjectFactory.create(1, {
			owner_id: currentUser.id,
			title: 'Colored Project',
			hex_color: 'e8590c',
		}))[0]
		await createDefaultViews(project.id)

		// A touchscreen desktop reports touch points while keeping a fine pointer and hover.
		// Playwright's hasTouch option can't model that – it also switches the pointer to coarse.
		await page.addInitScript(() => {
			Object.defineProperty(navigator, 'maxTouchPoints', {get: () => 1})
		})
		await page.goto('/')

		const item = page.locator(`li[data-project-id="${project.id}"]`)
		const colorBubble = item.locator('.color-bubble')
		const dragHandle = item.locator('.drag-handle')

		await expect(colorBubble).toBeVisible()
		expect(await page.evaluate(() => window.matchMedia('(pointer: fine)').matches)).toBe(true)
		expect(await page.evaluate(() => navigator.maxTouchPoints)).toBe(1)

		await expect(dragHandle).toHaveCSS('opacity', '0')
		await expect(colorBubble).toHaveCSS('opacity', '1')

		await item.locator('.list-menu-link').hover()

		await expect(dragHandle).toHaveCSS('opacity', '1')
		await expect(colorBubble).toHaveCSS('opacity', '0')
	})
})
