
export function updateUserSettings(settings) {
	const token = `Bearer ${window.localStorage.getItem('token')}`
	
	return cy.request({
		method: 'GET',
		url: `${Cypress.env('API_URL')}/user`,
		headers: {
			'Authorization': token,
		},
	})
		.its('body')
		.then(oldSettings => {
			return cy.request({
				method: 'POST',
				url: `${Cypress.env('API_URL')}/user/settings/general`,
				headers: {
					'Authorization': token,
				},
				body: {
					...oldSettings,
					...settings,
				},
			})
		})
}
