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
		const newNamespaceTitle = 'New Namespace'

		cy.visit('/namespaces')
		cy.get('a.button')
			.contains('Create a new namespace')
			.click()

		cy.url()
			.should('contain', '/namespaces/new')
		cy.get('.card-header-title')
			.should('contain', 'Create a new namespace')
		cy.get('input.input')
			.type(newNamespaceTitle)
		cy.get('.button')
			.contains('Create')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.namespace-container')
			.should('contain', newNamespaceTitle)
		cy.url()
			.should('contain', '/namespaces')
	})

	it('Should rename the namespace all places', () => {
		const newNamespaces = NamespaceFactory.create(5)
		const newNamespaceName = 'New namespace name'

		cy.visit('/namespaces')

		cy.get(`.namespace-container .menu.namespaces-lists .namespace-title:contains(${newNamespaces[0].title}) .dropdown .dropdown-trigger`)
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .namespace-title .dropdown .dropdown-content')
			.contains('Edit')
			.click()
		cy.url()
			.should('contain', '/settings/edit')
		cy.get('#namespacetext')
			.invoke('val')
			.should('equal', newNamespaces[0].title) // wait until the namespace data is loaded
		cy.get('#namespacetext')
			.type(`{selectall}${newNamespaceName}`)
		cy.get('footer.modal-card-foot .button')
			.contains('Save')
			.click()

		cy.get('.global-notification', { timeout: 1000 })
			.should('contain', 'Success')
		cy.get('.namespace-container .menu.namespaces-lists')
			.should('contain', newNamespaceName)
			.should('not.contain', newNamespaces[0].title)
		cy.get('.content.namespaces-list')
			.should('contain', newNamespaceName)
			.should('not.contain', newNamespaces[0].title)
	})

	it('Should remove a namespace when deleting it', () => {
		const newNamespaces = NamespaceFactory.create(5)

		cy.visit('/')

		cy.get(`.namespace-container .menu.namespaces-lists .namespace-title:contains(${newNamespaces[0].title}) .dropdown .dropdown-trigger`)
			.click()
		cy.get('.namespace-container .menu.namespaces-lists .namespace-title .dropdown .dropdown-content')
			.contains('Delete')
			.click()
		cy.url()
			.should('contain', '/settings/delete')
		cy.get('[data-cy="modalPrimary"]')
			.contains('Do it')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.namespace-container .menu.namespaces-lists')
			.should('not.contain', newNamespaces[0].title)
	})

	it('Should not show archived lists & namespaces if the filter is not checked', () => {
		const n = NamespaceFactory.create(1, {
			id: 2,
			is_archived: true,
		}, false)
		ListFactory.create(1, {
			id: 2,
			namespace_id: n[0].id,
		}, false)

		ListFactory.create(1, {
			id: 3,
			is_archived: true,
		}, false)

		// Initial
		cy.visit('/namespaces')
		cy.get('.namespaces-list .namespace')
			.should('not.contain', 'Archived')

		// Show archived
		cy.get('.namespaces-list .fancycheckbox.show-archived-check label.check span')
			.should('be.visible')
			.click()
		cy.get('.namespaces-list .fancycheckbox.show-archived-check input')
			.should('be.checked')
		cy.get('.namespaces-list .namespace')
			.should('contain', 'Archived')

		// Don't show archived
		cy.get('.namespaces-list .fancycheckbox.show-archived-check label.check span')
			.should('be.visible')
			.click()
		cy.get('.namespaces-list .fancycheckbox.show-archived-check input')
			.should('not.be.checked')

		// Second time visiting after unchecking
		cy.visit('/namespaces')
		cy.get('.namespaces-list .fancycheckbox.show-archived-check input')
			.should('not.be.checked')
		cy.get('.namespaces-list .namespace')
			.should('not.contain', 'Archived')
	})
})
