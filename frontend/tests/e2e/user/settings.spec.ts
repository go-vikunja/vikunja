import {test, expect} from '../../support/fixtures'

test.describe('User Settings', () => {
	// FIXME: File upload and image cropping functionality not working - upload response timeout
	test.skip('Changes the user avatar', async ({authenticatedPage: page}) => {
		const uploadAvatarPromise = page.waitForResponse(response =>
			response.url().includes('/user/settings/avatar/upload') && response.request().method() === 'POST',
		)

		await page.goto('/user/settings/avatar')

		await page.locator('input[name=avatarProvider][value=upload]').click()
		await page.locator('input[type=file]').setInputFiles('tests/fixtures/image.jpg')

		// Simulate the crop handler drag
		const handler = page.locator('.vue-handler-wrapper.vue-handler-wrapper--south .vue-simple-handler.vue-simple-handler--south')
		await handler.dispatchEvent('mousedown', {which: 1})
		await handler.dispatchEvent('mousemove', {clientY: 100})
		await handler.dispatchEvent('mouseup')

		await page.locator('[data-cy="uploadAvatar"]').filter({hasText: 'Upload Avatar'}).click()

		await uploadAvatarPromise
		await expect(page.locator('.global-notification')).toContainText('Success')
	})

	test('Updates the name', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/general')

		await page.locator('.general-settings input.input').first().fill('Lorem Ipsum')
		await page.locator('[data-cy="saveGeneralSettings"]').filter({hasText: 'Save'}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.navbar .username-dropdown-trigger .username')).toContainText('Lorem Ipsum')
	})
})
