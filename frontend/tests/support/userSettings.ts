import type {Page} from '@playwright/test'

export async function gotoUserSettings(page: Page, section: string) {
	await page.goto(`/user/settings/${section}`)
	await page.waitForLoadState('networkidle')
}
