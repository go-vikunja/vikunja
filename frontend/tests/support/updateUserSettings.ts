import type {APIRequestContext} from '@playwright/test'
import {objectToSnakeCase} from '../../src/helpers/case'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: any) {
	const apiUrl = process.env.API_URL || 'http://localhost:3456/api/v1'

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})

	const oldSettings = await userResponse.json()

	const snakeSettings = objectToSnakeCase(settings)

	// Deep merge frontend_settings if provided
	const mergedSettings = {
		...oldSettings,
		...snakeSettings,
	}

	if (snakeSettings.frontend_settings) {
		mergedSettings.frontend_settings = {
			...oldSettings.frontend_settings,
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
