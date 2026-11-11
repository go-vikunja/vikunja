import {test, expect} from '../../support/fixtures'
import {createProjects} from '../project/prepareProjects'

test.describe('Filter input focus', () => {
	test('only the wrapper shows a focus indicator', async ({authenticatedPage: page}) => {
		await createProjects()
		await page.goto('/projects/1/1')

		await page.locator('.filter-container button').filter({hasText: 'Filters'}).click()
		const editor = page.locator('.filter-popup .filter-input .ProseMirror')
		await editor.click()

		await expect(editor).toBeFocused()
		await expect(editor).toHaveCSS('box-shadow', 'none')
	})
})
