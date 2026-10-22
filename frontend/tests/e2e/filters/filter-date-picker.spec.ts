import {test, expect} from '../../support/fixtures'

test.describe('Filter Date Picker', () => {
	test('opens on a date value, moves to the next one and closes on an outside click', async ({authenticatedPage: page}) => {
		await page.goto('/filters/new')

		const filterInput = page.locator('.filter-input .ProseMirror')
		await expect(filterInput).toBeVisible()
		await filterInput.click()
		await page.keyboard.press('ControlOrMeta+a')
		await page.keyboard.press('Backspace')
		await filterInput.pressSequentially('dueDate = 2024-01-01 && startDate = 2024-02-02', {delay: 20})

		const dateValues = page.locator('.filter-input .date-value')
		await expect(dateValues).toHaveCount(2)

		const picker = page.locator('.filter-datepicker .datepicker-with-range')
		await expect(picker).not.toBeVisible()

		await dateValues.first().click()
		await expect(picker).toBeVisible()

		// The click on the second value light-dismisses the picker; it has to come back on the new anchor.
		await dateValues.nth(1).click()
		await expect(picker).toBeVisible()

		await page.locator('input#Title').click()
		await expect(picker).not.toBeVisible()
	})
})
