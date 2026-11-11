import type {APIRequestContext} from '@playwright/test'
import {apiV1Url} from './apiUrl'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: any) {
	const apiUrl = apiV1Url()

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})
	if (!userResponse.ok()) {
		throw new Error(`updateUserSettings: GET ${apiUrl}/user failed with ${userResponse.status()}`)
	}

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

	const updateResponse = await apiContext.post(`${apiUrl}/user/settings/general`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
		data: mergedSettings,
	})
	// Silence here means callers assert against settings that never changed.
	if (!updateResponse.ok()) {
		throw new Error(`updateUserSettings: POST failed with ${updateResponse.status()}: ${await updateResponse.text()}`)
	}
}
