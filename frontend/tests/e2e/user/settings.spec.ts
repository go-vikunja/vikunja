import {test, expect} from '../../support/fixtures'

test.describe('User Settings', () => {
	test('Changes the user avatar', async ({authenticatedPage: page, apiContext}) => {
		// This test uses the API directly to upload the avatar since the vue-advanced-cropper
		// component's canvas.toBlob() can return null in headless browser environments
		await page.goto('/user/settings/avatar')
		await page.waitForLoadState('networkidle')

		// Get the auth token from localStorage
		const token = await page.evaluate(() => localStorage.getItem('token'))

		// Upload the avatar directly via API
		const fs = await import('fs')
		const path = await import('path')
		const fileBuffer = fs.readFileSync(path.join(process.cwd(), 'tests/fixtures/image.jpg'))

		const response = await apiContext.put('user/settings/avatar/upload', {
			multipart: {
				avatar: {
					name: 'avatar.jpg',
					mimeType: 'image/jpeg',
					buffer: fileBuffer,
				},
			},
			headers: {
				'Authorization': `Bearer ${token}`,
			},
		})

		expect(response.ok()).toBe(true)

		// Reload the page to verify the avatar was updated
		await page.reload()
		await page.waitForLoadState('networkidle')

		// Verify the upload radio is now checked (indicating upload provider is set)
		const uploadRadio = page.locator('input[name=avatarProvider][value=upload]')
		await expect(uploadRadio).toBeChecked()
	})

	test('Updates the name', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/general')
		await page.waitForLoadState('networkidle')

		// Wait for the settings page to be fully loaded and the input to be enabled
		const nameInput = page.locator('.general-settings input.input').first()
		await expect(nameInput).toBeVisible({timeout: 10000})
		await expect(nameInput).toBeEnabled()

		// Clear and type to ensure Vue's reactivity is triggered
		await nameInput.clear()
		await nameInput.pressSequentially('Lorem Ipsum', {delay: 10})

		// The save button only appears when isDirty becomes true (settings changed)
		const saveButton = page.locator('[data-cy="saveGeneralSettings"]')
		await expect(saveButton).toBeVisible({timeout: 10000})
		await saveButton.click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.locator('.navbar .username-dropdown-trigger .username')).toContainText('Lorem Ipsum')
	})
})
