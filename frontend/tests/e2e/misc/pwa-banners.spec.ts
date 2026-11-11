import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {UserFactory} from '../../factories/user'
import {TaskFactory} from '../../factories/task'
import {createProjects} from '../project/prepareProjects'
import {setupApiUrl} from '../../support/authenticateUser'
import {TEST_PASSWORD} from '../../support/constants'

// The add to home screen banner is hidden from tablet width up
const iPhone8 = {
	width: 375,
	height: 667,
}

test.describe('PWA banners', () => {
	test.use({viewport: iPhone8})

	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
		await page.goto('/login')
		await expect(page.locator('input[id=username]')).toBeVisible()
		// Service workers are blocked in e2e, so fake the event registerServiceWorker.ts dispatches
		await page.evaluate(() => document.dispatchEvent(new CustomEvent('swUpdated', {detail: {waiting: null}})))
	})

	test('Are hidden on the login page', async ({page}) => {
		await expect(page.locator('.update-notification')).toBeHidden()
		await expect(page.locator('.add-to-home-screen')).toBeHidden()
	})

	test('Show after logging in, including an update found on the login page', async ({page}) => {
		const [user] = await UserFactory.create(1)
		await page.locator('input[id=username]').fill(user.username)
		await page.locator('input[id=password]').fill(TEST_PASSWORD)
		await page.locator('.button').filter({hasText: 'Login'}).click()
		await expect(page).toHaveURL('/')

		await expect(page.locator('.update-notification')).toBeVisible()
		await expect(page.locator('.add-to-home-screen')).toBeVisible()
	})

	test('Stack without overlapping each other or toasts', async ({page}) => {
		await showToastWithUpdateBanner(page)

		const toast = page.locator('.vue-notification.success')
		const install = page.locator('.add-to-home-screen')
		const update = page.locator('.update-notification')
		await expect(toast).toBeVisible()
		await expect(install).toBeVisible()
		await expect(update).toBeVisible()

		const [toastBox, installBox, updateBox] = await Promise.all([toast, install, update].map(l => l.boundingBox()))
		expect(toastBox!.y + toastBox!.height).toBeLessThanOrEqual(installBox!.y)
		expect(installBox!.y + installBox!.height).toBeLessThanOrEqual(updateBox!.y)
	})
})

test.describe('PWA banners on desktop', () => {
	test.use({viewport: {width: 1280, height: 800}})

	test.beforeEach(async ({page}) => {
		await setupApiUrl(page)
		await page.goto('/login')
		await expect(page.locator('input[id=username]')).toBeVisible()
	})

	test('Keep the update banner centered without lifting the toasts', async ({page}) => {
		await showToastWithUpdateBanner(page)

		const toast = page.locator('.vue-notification.success')
		const update = page.locator('.update-notification')
		await expect(toast).toBeVisible()
		await expect(update).toBeVisible()

		const [toastBox, updateBox] = await Promise.all([toast, update].map(l => l.boundingBox()))
		expect(Math.abs(updateBox!.x + updateBox!.width / 2 - 640)).toBeLessThan(2)
		// Side by side, so both share the bottom margin.
		expect(Math.abs((toastBox!.y + toastBox!.height) - (updateBox!.y + updateBox!.height))).toBeLessThan(2)
	})
})

async function showToastWithUpdateBanner(page: Page) {
	const [user] = await UserFactory.create(1)
	await createProjects(1)
	await TaskFactory.create(1, {project_id: 1})
	await page.locator('input[id=username]').fill(user.username)
	await page.locator('input[id=password]').fill(TEST_PASSWORD)
	await page.locator('.button').filter({hasText: 'Login'}).click()
	await expect(page).toHaveURL('/')

	await page.goto('/projects/1/1')
	await expect(page.locator('.tasks .task')).toBeVisible()
	await page.evaluate(() => document.dispatchEvent(new CustomEvent('swUpdated', {detail: {waiting: null}})))
	await page.locator('.tasks .task .fancy-checkbox').first().click()
}
