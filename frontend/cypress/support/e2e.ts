
import './commands'
import '@4tw/cypress-drag-drop'

// see https://github.com/cypress-io/cypress/issues/702#issuecomment-587127275
Cypress.on('window:before:load', (win) => {
	// disable service workers
	// @ts-expect-error __proto__ has no types
	delete win.navigator.__proto__.ServiceWorker
})