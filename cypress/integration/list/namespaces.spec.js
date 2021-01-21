import {UserFactory} from '../../factories/user'

import '../../support/authenticateUser'
import {ListFactory} from '../../factories/list'
import {NamespaceFactory} from '../../factories/namespace'

describe('Namepaces', () => {
	let namespaces

	beforeEach(() => {
		UserFactory.create(1)
		namespaces = NamespaceFactory.create(1)
		ListFactory.create(1)
	})

	it('Should be all there', () => {
		cy.visit('/namespaces')
		cy.get('.namespace h1 span')
			.should('contain', namespaces[0].title)
	})

	it('Should create a new Namespace', () => {
		cy.visit('/namespaces')
		cy.get('a.button')
			.contains('Create namespace')
			.click()
		cy.url()
			.should('contain', '/namespaces/new')
		cy.get('.card-header-title')
			.should('contain', 'Create a new namespace')
		cy.get('input.input')
			.type('New Namespace')
		cy.get('.button')
			.contains('Create')
			.click()
		cy.url()
			.should('contain', '/namespaces')
	})
})
