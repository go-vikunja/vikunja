/// <reference types="cypress" />
// ***********************************************
// This example commands.ts shows you how to
// create various custom commands and overwrite
// existing commands.
//
// For more comprehensive examples of custom
// commands please read more here:
// https://on.cypress.io/custom-commands
// ***********************************************
//
//
// -- This is a parent command --
// Cypress.Commands.add('login', (email, password) => { ... })
//
//
// -- This is a child command --
// Cypress.Commands.add('drag', { prevSubject: 'element'}, (subject, options) => { ... })
//
//
// -- This is a dual command --
// Cypress.Commands.add('dismiss', { prevSubject: 'optional'}, (subject, options) => { ... })
//
//
// -- This will overwrite an existing command --
// Cypress.Commands.overwrite('visit', (originalFn, url, options) => { ... })
//
// declare global {
//   namespace Cypress {
//     interface Chainable {
//       login(email: string, password: string): Chainable<void>
//       drag(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       dismiss(subject: string, options?: Partial<TypeOptions>): Chainable<Element>
//       visit(originalFn: CommandOriginalFn, url: string, options: Partial<VisitOptions>): Chainable<Element>
//     }
//   }
// }

Cypress.Commands.add('pasteFile', {prevSubject: true}, (subject, fileName, fileType = 'image/png') => {
	// Load the file fixture as base64
	cy.fixture(fileName, 'base64').then((fileContent) => {
		// Convert base64 to a Blob
		const blob = Cypress.Blob.base64StringToBlob(fileContent, fileType)
		// Create a File object
		const testFile = new File([blob], fileName, {type: fileType})
		// Create a DataTransfer and add the file
		const dataTransfer = new DataTransfer()
		dataTransfer.items.add(testFile)

		// Create the paste event with clipboardData containing the file
		const pasteEvent = new ClipboardEvent('paste', {
			bubbles: true,
			cancelable: true,
			clipboardData: dataTransfer,
		})

		// Dispatch the paste event on the target element
		subject[0].dispatchEvent(pasteEvent)
	})
})

