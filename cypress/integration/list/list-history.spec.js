import {ListFactory} from '../../factories/list'

import '../../support/authenticateUser'

describe('List History', () => {
	it('should show a list history on the home page', () => {
		cy.intercept(Cypress.env('API_URL') + '/namespaces*').as('loadNamespaces')
		
		const lists = ListFactory.create(6)

		cy.visit('/')
		cy.wait('@loadNamespaces')
		cy.get('h3')
			.contains('Last viewed')
			.should('not.exist')

		cy.visit(`/lists/${lists[0].id}`)
		cy.wait('@loadNamespaces')
		cy.visit(`/lists/${lists[1].id}`)
		cy.wait('@loadNamespaces')
		cy.visit(`/lists/${lists[2].id}`)
		cy.wait('@loadNamespaces')
		cy.visit(`/lists/${lists[3].id}`)
		cy.wait('@loadNamespaces')
		cy.visit(`/lists/${lists[4].id}`)
		cy.wait('@loadNamespaces')
		cy.visit(`/lists/${lists[5].id}`)
		cy.wait('@loadNamespaces')

		cy.visit('/')
		cy.wait('@loadNamespaces')
		
		cy.get('h3')
			.contains('Last viewed')
			.should('exist')
		cy.get('.list-cards-wrapper-2-rows')
			.should('not.contain', lists[0].title)
			.should('contain', lists[1].title)
			.should('contain', lists[2].title)
			.should('contain', lists[3].title)
			.should('contain', lists[4].title)
			.should('contain', lists[5].title)
	})
})