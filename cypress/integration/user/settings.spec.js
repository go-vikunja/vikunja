import {UserFactory} from '../../factories/user'

import '../../support/authenticateUser'

describe('User Settings', () => {
	beforeEach(() => {
		UserFactory.create(1)
	})

	it('Changes the user avatar', () => {
		cy.visit('/user/settings')

		cy.get('input[name=avatarProvider][value=upload]')
			.click()
		cy.get('input[type=file]')
			.attachFile('image.jpg')
		cy.get('.vue-handler-wrapper.vue-handler-wrapper--south .vue-simple-handler.vue-simple-handler--south')
			.trigger('mousedown', {which: 1})
			.trigger('mousemove', {clientY: 100})
			.trigger('mouseup')
		cy.get('a.button.is-primary')
			.contains('Upload Avatar')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
	})

	it('Updates the name', () => {
		cy.visit('/user/settings')

		cy.get('input#newName')
			.type('Lorem Ipsum')
		cy.get('.card.update-name button.button.is-primary')
			.contains('Save')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.navbar .user .username')
			.should('contain', 'Lorem Ipsum')
	})
})
