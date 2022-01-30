import {ListFactory} from '../../factories/list'

import '../../support/authenticateUser'

describe('List History', () => {
	it('should show a list history on the home page', () => {
		cy.intercept(Cypress.env('API_URL') + '/namespaces*').as('loadNamespaces')
		cy.intercept(Cypress.env('API_URL') + '/lists/*').as('loadList')
		
		const lists = ListFactory.create(6)

		cy.visit('/')
		cy.wait('@loadNamespaces')
		cy.get('body')
			.should('not.contain', 'Last viewed')

		cy.visit(`/lists/${lists[0].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')
		cy.visit(`/lists/${lists[1].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')
		cy.visit(`/lists/${lists[2].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')
		cy.visit(`/lists/${lists[3].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')
		cy.visit(`/lists/${lists[4].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')
		cy.visit(`/lists/${lists[5].id}`)
		cy.wait('@loadNamespaces')
		cy.wait('@loadList')

		// cy.visit('/')
		// cy.wait('@loadNamespaces')
		// Not using cy.visit here to work around the redirect issue fixed in #1337
		cy.get('nav.menu.top-menu a')
			.contains('Overview')
			.click()
		
		cy.get('body')
			.should('contain', 'Last viewed')
		cy.get('.list-cards-wrapper-2-rows')
			.should('not.contain', lists[0].title)
			.should('contain', lists[1].title)
			.should('contain', lists[2].title)
			.should('contain', lists[3].title)
			.should('contain', lists[4].title)
			.should('contain', lists[5].title)
	})
})