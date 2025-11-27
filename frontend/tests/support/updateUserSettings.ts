import type {APIRequestContext} from '@playwright/test'

export async function updateUserSettings(apiContext: APIRequestContext, token: string, settings: Record<string, unknown>) {
	const apiUrl = process.env.API_URL || 'http://localhost:3456/api/v1'

	const userResponse = await apiContext.get(`${apiUrl}/user`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
	})

	const oldSettings = await userResponse.json()

	await apiContext.post(`${apiUrl}/user/settings/general`, {
		headers: {
			'Authorization': `Bearer ${token}`,
		},
		data: {
			...oldSettings,
			...settings,
		},
	})
}
