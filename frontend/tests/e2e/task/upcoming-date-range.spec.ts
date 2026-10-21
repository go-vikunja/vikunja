import {test, expect} from '../../support/fixtures'
import type {Page} from '@playwright/test'

function dateKey(date: Date) {
	const month = String(date.getMonth() + 1).padStart(2, '0')
	const day = String(date.getDate()).padStart(2, '0')
	return `${date.getFullYear()}-${month}-${day}`
}

async function pickPreset(page: Page, name: string) {
	const popup = page.locator('.datepicker-with-range-container .popup')
	const trigger = page.getByRole('button', {name: 'Select a date range'})

	await trigger.click()
	await popup.locator('.selections').getByRole('button', {name, exact: true}).click()
	await expect(page).toHaveTitle(`${name} | Vikunja`)
	if (!await popup.isVisible()) {
		await trigger.click()
	}

	return popup.locator('.calendar-month__cell:is(.is-selected, .is-in-range) button')
}

test.describe('Upcoming date range picker', () => {
	test('highlights a datemath preset in the calendar', async ({authenticatedPage: page}) => {
		await page.goto('/tasks/by/upcoming')

		const today = new Date()
		const monday = new Date(today.getFullYear(), today.getMonth(), today.getDate() - (today.getDay() + 6) % 7)
		const week = Array.from({length: 7}, (_, i) => dateKey(new Date(monday.getFullYear(), monday.getMonth(), monday.getDate() + i)))

		const thisWeek = await pickPreset(page, 'This Week')
		await expect(thisWeek).toHaveCount(7)
		await expect(thisWeek.first()).toHaveAttribute('data-date', week[0])
		await expect(thisWeek.last()).toHaveAttribute('data-date', week[6])

		await page.keyboard.press('Escape')

		// The end of `now/d+1d` is exclusive, so tomorrow must not be highlighted.
		const todayRange = await pickPreset(page, 'Today')
		await expect(todayRange).toHaveCount(1)
		await expect(todayRange).toHaveAttribute('data-date', dateKey(today))
	})
})
