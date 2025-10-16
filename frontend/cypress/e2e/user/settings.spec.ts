import {createFakeUserAndLogin} from '../../support/authenticateUser'

describe('User Settings', () => {
	createFakeUserAndLogin()

	it('Changes the user avatar', () => {
		cy.visit('/user/settings/avatar')

		cy.intercept('PUT', '**/user/settings/avatar/upload').as('uploadAvatar')

		cy.get('input[name=avatarProvider][value=upload]')
			.click()
		cy.get('input[type=file]', {timeout: 1000})
			.selectFile('cypress/fixtures/image.jpg', {force: true}) // The input is not visible, but on purpose
		cy.get('.vue-handler-wrapper.vue-handler-wrapper--south .vue-simple-handler.vue-simple-handler--south')
			.trigger('mousedown', {which: 1})
			.trigger('mousemove', {clientY: 100})
			.trigger('mouseup')
		cy.get('[data-cy="uploadAvatar"]')
			.contains('Upload Avatar')
			.click()

		cy.wait('@uploadAvatar', {timeout: 10000})
		cy.get('.global-notification')
			.should('contain', 'Success')
	})

	it('Updates the name', () => {
		cy.visit('/user/settings/general')

		cy.get('.general-settings input.input')
			.first()
			.type('Lorem Ipsum')
		cy.get('[data-cy="saveGeneralSettings"]')
			.contains('Save')
			.click()

		cy.get('.global-notification')
			.should('contain', 'Success')
		cy.get('.navbar .username-dropdown-trigger .username')
			.should('contain', 'Lorem Ipsum')
	})
})
