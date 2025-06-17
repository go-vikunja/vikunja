
import './commands'
import '@4tw/cypress-drag-drop'

beforeEach(() => {
	// More comprehensive intercept for all API requests to dev server
	cy.intercept('GET', '**/api/v1/**', (req) => {
		// If the request is going to the dev server, redirect to API server
		if (req.url.includes('127.0.0.1:4173') || req.url.includes('localhost:4173')) {
			const newUrl = req.url.replace(/https?:\/\/(127\.0\.0\.1|localhost):4173\/api\/v1/, Cypress.env('API_URL'))
			console.log('Redirecting request from', req.url, 'to', newUrl)
			req.url = newUrl
		}
	}).as('apiRequest')
})

// see https://github.com/cypress-io/cypress/issues/702#issuecomment-587127275
Cypress.on('window:before:load', (win) => {
	// disable service workers
	// @ts-ignore
	delete win.navigator.__proto__.ServiceWorker
})