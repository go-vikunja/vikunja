import {test, expect} from '../../support/fixtures'

test.describe('User Settings', () => {
	test('Changes the user avatar', async ({authenticatedPage: page}) => {
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

		// Wait for the cropper to be ready (button becomes enabled when canvas is ready)
		await expect(uploadButton).toBeEnabled({timeout: 10000})

		// Set up response waiter before clicking
		const avatarUploadPromise = page.waitForResponse(response =>
			response.url().includes('avatar') && response.request().method() === 'PUT',
		)

		await uploadButton.click()

		// Wait for the avatar upload response and verify it succeeded
		const response = await avatarUploadPromise
		expect(response.ok()).toBe(true)

		await expect(page.locator('.global-notification')).toContainText('Success', {timeout: 10000})
	})

	test('Invalidates avatar cache when uploading a new avatar', async ({authenticatedPage: page}) => {
		await page.goto('/user/settings/avatar')
		await page.waitForLoadState('networkidle')

		const uploadRadio = page.locator('input[name=avatarProvider][value=upload]')
		await expect(uploadRadio).toBeVisible({timeout: 5000})
		await uploadRadio.click()

		const fileInput = page.locator('input[type=file]')
		const uploadButton = page.locator('[data-cy="uploadAvatar"]')
		const headerAvatar = page.locator('.username-dropdown-trigger img.avatar')
		const notification = page.locator('.global-notification')

		// Upload first avatar (image.jpg)
		await fileInput.setInputFiles('tests/fixtures/image.jpg')
		await expect(uploadButton).toBeEnabled({timeout: 10000})

		const firstUploadPromise = page.waitForResponse(response =>
			response.url().includes('avatar') && response.request().method() === 'PUT',
		)
		await uploadButton.click()
		const firstResponse = await firstUploadPromise
		expect(firstResponse.ok()).toBe(true)
		await expect(notification).toContainText('Success', {timeout: 10000})

		// Wait for the header avatar to update and capture its src
		await expect(headerAvatar).toHaveAttribute('src', /blob:|data:/, {timeout: 10000})
		const firstAvatarSrc = await headerAvatar.getAttribute('src')

		// Wait for the notification to disappear before uploading again
		await expect(notification).not.toBeVisible({timeout: 10000})

		// Upload second avatar (image-blue.png)
		await fileInput.setInputFiles('tests/fixtures/image-blue.png')
		await expect(uploadButton).toBeEnabled({timeout: 10000})

		const secondUploadPromise = page.waitForResponse(response =>
			response.url().includes('avatar') && response.request().method() === 'PUT',
		)
		await uploadButton.click()
		const secondResponse = await secondUploadPromise
		expect(secondResponse.ok()).toBe(true)
		await expect(notification).toContainText('Success', {timeout: 10000})

		// Verify the header avatar changed to a different blob URL
		await expect(headerAvatar).not.toHaveAttribute('src', firstAvatarSrc!, {timeout: 10000})
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
