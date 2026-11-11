import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

// https://github.com/go-vikunja/vikunja/issues/4008
test.describe('Editor double-click to edit', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	test('keeps the clicked text in place and puts the cursor there', async ({authenticatedPage: page}) => {
		const paragraphs = Array.from({length: 20}, (_, i) => `<p>paragraph${i} alpha beta gamma delta</p>`)
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: paragraphs.join(''),
		})

		await page.setViewportSize({width: 1280, height: 600})
		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const description = page.locator('.task-view .details.content.description')
		const target = description.locator('.ProseMirror p').nth(5)
		await target.scrollIntoViewIfNeeded()

		const topBefore = (await target.boundingBox())!.y
		await target.dblclick({position: {x: 5, y: 5}})

		await expect(description.locator('.editor-toolbar')).toBeVisible()
		expect(Math.abs((await target.boundingBox())!.y - topBefore)).toBeLessThan(2)

		await page.keyboard.type('X')
		await expect(target).toContainText('X')
		await expect(description.locator('.ProseMirror p').first()).not.toContainText('X')
	})
})
