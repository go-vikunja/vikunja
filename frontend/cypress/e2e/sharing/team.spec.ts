import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TeamFactory} from '../../factories/team'
import {TeamMemberFactory} from '../../factories/team_member'
import {UserFactory} from '../../factories/user'

describe('Team', () => {
	createFakeUserAndLogin()

    it('Creates a new team', () => {
        TeamFactory.truncate()
        cy.visit('/teams')

        const newTeamName = 'New Team'

        cy.get('a.button')
            .contains('Create a team')
            .click()
		cy.url()
			.should('contain', '/teams/new')
		cy.get('.card-header-title')
			.contains('Create a team')
		cy.get('input.input')
			.type(newTeamName)
		cy.get('.button')
			.contains('Create')
			.click()

		cy.url()
            .should('contain', '/edit')
        cy.get('input#teamtext')
            .should('have.value', newTeamName)
    })

    it('Shows all teams', () => {
        TeamMemberFactory.create(10, {
            team_id: '{increment}',
        })
        const teams = TeamFactory.create(10, {
            id: '{increment}',
        })

        cy.visit('/teams')

        cy.get('.teams.box')
            .should('not.be.empty')
        teams.forEach(t => {
            cy.get('.teams.box')
                .should('contain', t.name)
        })
    })

    it('Allows an admin to edit the team', () => {
        TeamMemberFactory.create(1, {
            team_id: 1,
            admin: true,
        })
        const teams = TeamFactory.create(1, {
            id: 1,
        })
        
        cy.visit('/teams/1/edit')
        cy.get('.card input.input')
            .first()
            .type('{selectall}New Team Name')

        cy.get('.card .button')
            .contains('Save')
            .click()
        
        cy.get('table.table td')
            .contains('Admin')
            .should('exist')
        cy.get('.global-notification')
			.should('contain', 'Success')
    })

    it('Does not allow a normal user to edit the team', () => {
        TeamMemberFactory.create(1, {
            team_id: 1,
            admin: false,
        })
        const teams = TeamFactory.create(1, {
            id: 1,
        })
        
        cy.visit('/teams/1/edit')
        cy.get('.card input.input')
            .should('not.exist')
        cy.get('table.table td')
            .contains('Member')
            .should('exist')
    })

    it('Allows an admin to add members to the team', () => {
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
