import {test, expect} from '../../support/fixtures'
import type {CDPSession} from '@playwright/test'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

// Android keyboards keep the current word in an IME composition and report every key as 229.
async function composeText(cdp: CDPSession, text: string) {
	for (let i = 1; i <= text.length; i++) {
		await cdp.send('Input.dispatchKeyEvent', {type: 'rawKeyDown', windowsVirtualKeyCode: 229})
		await cdp.send('Input.imeSetComposition', {text: text.slice(0, i), selectionStart: i, selectionEnd: i})
		await cdp.send('Input.dispatchKeyEvent', {type: 'keyUp', windowsVirtualKeyCode: 229})
	}
}

test.describe('Quick actions with an IME keyboard', () => {
	test.use({viewport: {width: 412, height: 915}, hasTouch: true})

	test('Opens the tapped task while the search word is still being composed', async ({authenticatedPage: page}) => {
		const [project] = await ProjectFactory.create(1)
		await createDefaultViews(project.id)
		await TaskFactory.create(8, {
			project_id: project.id,
			title: i => i <= 4 ? `Other task ${i}` : `Milk task ${i}`,
		})

		await page.goto('/')
		await page.getByRole('button', {name: 'Open the search/quick action bar'}).click()
		const input = page.locator('.quick-actions .action-input input')
		await expect(input).toBeFocused()

		const cdp = await page.context().newCDPSession(page)
		await composeText(cdp, 'milk')

		const tasks = page.locator('.quick-actions .result-item-button').filter({hasText: 'Milk task'})
		await expect(tasks).toHaveCount(4)
		await expect(page.locator('.quick-actions .result-item-button')).toHaveCount(4)

		const box = await tasks.nth(1).boundingBox()
		await page.touchscreen.tap(box!.x + box!.width / 2, box!.y + box!.height / 2)

		await expect(page).toHaveURL(/\/tasks\/6$/)
	})
})
