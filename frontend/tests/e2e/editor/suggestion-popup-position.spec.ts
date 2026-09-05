import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews} from '../project/prepareProjects'

async function openDescriptionEditor(page) {
	const editButton = page.locator('.task-view .details.content.description .tiptap button.done-edit').filter({hasText: 'Edit'})
	await expect(editButton).toBeVisible({timeout: 10000})
	await editButton.click()

	const editor = page.locator('.task-view .details.content.description [contenteditable="true"]').first()
	await expect(editor).toBeVisible({timeout: 10000})
	return editor
}

test.describe('Suggestion popup positioning', () => {
	let projectId: number

	test.beforeEach(async () => {
		const projects = await ProjectFactory.create(1) as Array<{id: number}>
		projectId = projects[0].id
		const views = await createDefaultViews(projectId)
		await BucketFactory.create(1, {
			project_view_id: views[3].id,
		})
	})

	test('the popup is anchored to the suggestion text', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			project_id: projectId,
			description: 'x',
			index: 1,
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const editor = await openDescriptionEditor(page)
		await editor.click()
		await page.keyboard.press('ControlOrMeta+a')
		await page.keyboard.press('Delete')

		await editor.pressSequentially('/head', {delay: 60})

		const popup = page.locator('.items').first()
		await expect(popup).toBeVisible()
		await page.waitForTimeout(400)

		const anchor = await page.locator('[data-decoration-id]').first().boundingBox()
		const box = await popup.boundingBox()
		console.log('anchor:', JSON.stringify(anchor), 'popup:', JSON.stringify(box))

		expect(anchor).not.toBeNull()
		expect(box).not.toBeNull()
		expect(Math.abs(box!.x - anchor!.x)).toBeLessThan(20)
		expect(Math.abs(box!.y - (anchor!.y + anchor!.height))).toBeLessThan(30)
	})

	test('the popup follows the text when the page scrolls', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 2,
			project_id: projectId,
			description: '<p>filler line</p>'.repeat(40),
			index: 1,
		})

		await page.setViewportSize({width: 1280, height: 600})
		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const editor = await openDescriptionEditor(page)
		await editor.click()
		await page.keyboard.press('ControlOrMeta+End')
		await page.keyboard.press('Enter')

		await editor.pressSequentially('/head', {delay: 60})

		const popup = page.locator('.items').first()
		await expect(popup).toBeVisible()
		await page.waitForTimeout(400)

		const anchorBefore = await page.locator('[data-decoration-id]').first().boundingBox()
		const popupBefore = await popup.boundingBox()

		await editor.hover()
		await page.mouse.wheel(0, -200)
		await page.waitForTimeout(600)

		const anchorAfter = await page.locator('[data-decoration-id]').first().boundingBox()
		const popupAfter = await popup.boundingBox()

		console.log('before -> anchor:', JSON.stringify(anchorBefore), 'popup:', JSON.stringify(popupBefore))
		console.log('after  -> anchor:', JSON.stringify(anchorAfter), 'popup:', JSON.stringify(popupAfter))

		expect(anchorBefore).not.toBeNull()
		expect(anchorAfter).not.toBeNull()
		// The page has to actually scroll for this test to mean anything.
		expect(Math.abs(anchorAfter!.y - anchorBefore!.y)).toBeGreaterThan(50)

		const offsetBefore = popupBefore!.y - anchorBefore!.y
		const offsetAfter = popupAfter!.y - anchorAfter!.y
		expect(Math.abs(offsetAfter - offsetBefore)).toBeLessThan(15)
	})
})
