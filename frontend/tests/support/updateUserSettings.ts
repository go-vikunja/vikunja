import type {APIRequestContext} from '@playwright/test'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: any) {
	const apiUrl = process.env.API_URL || 'http://localhost:3456/api/v1'

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})

	const oldSettings = await userResponse.json()

	// Deep merge frontendSettings if provided
	const mergedSettings = {
		...oldSettings,
		...settings,
	}

	if (settings.frontendSettings) {
		mergedSettings.frontendSettings = {
			...oldSettings.frontendSettings,
			...settings.frontendSettings,
		}
	}

	await apiContext.post(`${apiUrl}/user/settings/general`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
		data: mergedSettings,
	})
}
