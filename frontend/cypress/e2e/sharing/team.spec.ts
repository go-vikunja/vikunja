import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {UserFactory} from '../../factories/user'

describe('Team', () => {
	createFakeUserAndLogin()

    it.skip('Allows an admin to add members to the team', () => {
        TeamMemberFactory.create(1, {
            team_id: 1,
            admin: true,
        })
        TeamFactory.create(1, {
            id: 1,
        })
        const users = UserFactory.create(5)
        
        cy.visit('/teams/1/edit')
        cy.get('.card')
            .contains('Team Members')
            .get('.card-content .multiselect .input-wrapper input')
            .type(users[1].username)
        cy.get('.card')
            .contains('Team Members')
            .get('.card-content .multiselect .search-results')
				.children()
				.first()
				.click()
        cy.get('.card')
            .contains('Team Members')
            .get('.card-content .button')
            .contains('Add to team')
            .click()
        
        cy.get('table.table td')
            .contains('Admin')
            .should('exist')
        cy.get('table.table tr')
            .should('contain', users[1].username)
            .should('contain', 'Member')
        cy.get('.global-notification')
			.should('contain', 'Success')
    })
})
