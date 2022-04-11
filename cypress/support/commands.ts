/**
 * Recursively gets an element, returning only after it's determined to be attached to the DOM for good.
 *
 * Source: https://github.com/cypress-io/cypress/issues/7306#issuecomment-850621378
 */
Cypress.Commands.add('getSettled', (selector, opts = {}) => {
	const retries = opts.retries || 3
	const delay = opts.delay || 100

	const isAttached = (resolve, count = 0) => {
		const el = Cypress.$(selector)

		// is element attached to the DOM?
		count = Cypress.dom.isAttached(el) ? count + 1 : 0

		// hit our base case, return the element
		if (count >= retries) {
			return resolve(el)
		}

		// retry after a bit of a delay
		setTimeout(() => isAttached(resolve, count), delay)
	}

	// wrap, so we can chain cypress commands off the result
	return cy.wrap(null).then(() => {
		return new Cypress.Promise((resolve) => {
			return isAttached(resolve, 0)
		}).then((el) => {
			return cy.wrap(el)
		})
	})
})
