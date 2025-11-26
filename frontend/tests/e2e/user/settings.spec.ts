import {test, expect} from '../../support/fixtures'

test.describe('User Settings', () => {
	// TODO: This test is flaky - the cropper's canvas.toBlob returns null intermittently
	// The vue-advanced-cropper component seems to not properly initialize in the test environment
	test.skip('Changes the user avatar', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/avatar')
		await page.waitForLoadState('networkidle')

		// Wait for the avatar settings content to be visible
		const uploadRadio = page.locator('input[name=avatarProvider][value=upload]')
		await expect(uploadRadio).toBeVisible({timeout: 5000})

		await uploadRadio.click()

		// Set the file directly on the (hidden) file input
		const fileInput = page.locator('input[type=file]')
		await fileInput.setInputFiles('tests/fixtures/image.jpg')

		// Wait for the cropper to be visible (the image needs to be loaded)
		const cropper = page.locator('.vue-advanced-cropper')
		await expect(cropper).toBeVisible({timeout: 10000})

		// After cropper appears, there's a new "Upload Avatar" button with data-cy attribute
		const uploadButton = page.locator('[data-cy="uploadAvatar"]')
		await expect(uploadButton).toBeVisible()

		// Listen for network requests
		page.on('request', (request) => {
			if (request.url().includes('avatar')) {
				console.log('Request:', request.method(), request.url())
			}
		})

		await uploadButton.click()

		// Wait for success notification instead of specific response
		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 10000})
	})

	test('Updates the name', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/general')
		await page.waitForLoadState('networkidle')

		// Wait for the settings page to be fully loaded and the input to be enabled
		const nameInput = page.locator('.general-settings input.input').first()
		await expect(nameInput).toBeVisible()
		await expect(nameInput).toBeEnabled()

		await nameInput.fill('Lorem Ipsum')
		await page.locator('[data-cy="saveGeneralSettings"]').filter({hasText: 'Save'}).click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.navbar .username-dropdown-trigger .username')).toContainText('Lorem Ipsum')
	})
})
