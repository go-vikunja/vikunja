import {LinkShareFactory} from '../../factories/link_sharing'
import {TaskFactory} from '../../factories/task'
import {UserFactory} from '../../factories/user'
import {createProjects} from '../project/prepareProjects'

function prepareLinkShare() {
	UserFactory.create()
	const projects = createProjects()
	const tasks = TaskFactory.create(10, {
		project_id: projects[0].id,
	})
	const linkShares = LinkShareFactory.create(1, {
		project_id: projects[0].id,
		permission: 0,
	})
	
	return {
		share: linkShares[0],
		project: projects[0],
		tasks,
	}
}

describe('Link shares', () => {
	it('Can view a link share', () => {
		const {share, project, tasks} = prepareLinkShare()

		// Set up comprehensive API intercepts for all possible task loading endpoints BEFORE navigation
		cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')

		cy.visit(`/share/${share.hash}/auth`)

		// Wait for redirect to complete
		cy.url().should('contain', `/projects/${project.id}/1#share-auth-token=${share.hash}`)

		// Wait for project title to load
		cy.get('h1.title')
			.should('be.visible')
			.should('contain', project.title)

		// Wait for tasks to load from API
		cy.wait('@loadTasks', {timeout: 30000})

		// Verify it's a read-only share (no task input)
		cy.get('input.input[placeholder="Add a task…"]')
			.should('not.exist')

		// Wait for tasks container to be visible and contain the task
		cy.get('.tasks')
			.should('be.visible')
			.and('contain', tasks[0].title)
	})

	it('Should work when directly viewing a project with share hash present', () => {
		const {share, project, tasks} = prepareLinkShare()

		// Set up comprehensive API intercepts for all possible task loading endpoints BEFORE navigation
		cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')

		cy.visit(`/projects/${project.id}/1#share-auth-token=${share.hash}`)

		// Wait for project title to load
		cy.get('h1.title')
			.should('be.visible')
			.should('contain', project.title)

		// Wait for tasks to load from API
		cy.wait('@loadTasks', {timeout: 30000})

		// Verify it's a read-only share (no task input)
		cy.get('input.input[placeholder="Add a task…"]')
			.should('not.exist')

		// Wait for tasks container to be visible and contain the task
		cy.get('.tasks')
			.should('be.visible')
			.and('contain', tasks[0].title)
	})
	
	it('Should work when directly viewing a task with share hash present', () => {
		const {share, project, tasks} = prepareLinkShare()

		cy.visit(`/tasks/${tasks[0].id}#share-auth-token=${share.hash}`)

		cy.get('h1.title')
			.should('contain', tasks[0].title)
	})
})
