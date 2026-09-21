import type {APIRequestContext} from '@playwright/test'
import {apiV1Url} from './apiUrl'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: any) {
	const apiUrl = apiV1Url()

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})

	const userData = await userResponse.json()
	// GET /user returns { settings: { frontend_settings: ... }, ... }
	// POST /user/settings/general expects { frontend_settings: ... } at the top level
	const oldSettings = userData.settings || {}

	// Deep merge frontend_settings if provided
	const mergedSettings = {
		...oldSettings,
		...settings,
	}

	if (settings.frontend_settings) {
		mergedSettings.frontend_settings = {
			...(oldSettings.frontend_settings || {}),
			...settings.frontend_settings,
		}
	}

	await apiContext.post(`${apiUrl}/user/settings/general`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
		data: mergedSettings,
	})
}
