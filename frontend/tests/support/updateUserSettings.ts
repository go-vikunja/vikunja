import type {APIRequestContext} from '@playwright/test'
import {objectToSnakeCase} from '../../src/helpers/case'
import {TEST_API_URL} from './constants'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: any) {
	const apiUrl = TEST_API_URL

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})

	const userData = await userResponse.json()
	// GET /user returns { settings: { frontend_settings: ... }, ... }
	// POST /user/settings/general expects { frontend_settings: ... } at the top level
	const oldSettings = userData.settings || {}

	const snakeSettings = objectToSnakeCase(settings)

	// Deep merge frontend_settings if provided
	const mergedSettings = {
		...oldSettings,
		...snakeSettings,
	}

	if (snakeSettings.frontend_settings) {
		mergedSettings.frontend_settings = {
			...(oldSettings.frontend_settings || {}),
			...snakeSettings.frontend_settings,
		}
	}

	await apiContext.post(`${apiUrl}/user/settings/general`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
		data: mergedSettings,
	})
}
