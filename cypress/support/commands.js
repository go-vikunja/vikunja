/**
 * getAttached(selector)
 * getAttached(selectorFn)
 *
 * Waits until the selector finds an attached element, then yields it (wrapped).
 * selectorFn, if provided, is passed $(document). Don't use cy methods inside selectorFn.
 *
 * Source: https://github.com/cypress-io/cypress/issues/5743#issuecomment-650421731
 */
Cypress.Commands.add('getAttached', selector => {
	const getElement = typeof selector === 'function' ? selector : $d => $d.find(selector);
	let $el = null;
	return cy.document().should($d => {
		$el = getElement(Cypress.$($d));
		expect(Cypress.dom.isDetached($el)).to.be.false;
	}).then(() => cy.wrap($el));
});
