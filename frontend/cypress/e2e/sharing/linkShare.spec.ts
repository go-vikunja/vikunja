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

		cy.visit(`/share/${share.hash}/auth`)

		cy.get('h1.title')
			.should('contain', project.title)
		cy.get('input.input[placeholder="Add a task…"]')
			.should('not.exist')
		cy.get('.tasks')
			.should('contain', tasks[0].title)
		
		cy.url().should('contain', `/projects/${project.id}/1#share-auth-token=${share.hash}`)
	})

	it('Should work when directly viewing a project with share hash present', () => {
		const {share, project, tasks} = prepareLinkShare()

		cy.visit(`/projects/${project.id}/1#share-auth-token=${share.hash}`)

		cy.get('h1.title')
			.should('contain', project.title)
		cy.get('input.input[placeholder="Add a task…"]')
			.should('not.exist')
		cy.get('.tasks')
			.should('contain', tasks[0].title)
	})
	
	it('Should work when directly viewing a task with share hash present', () => {
		const {share, project, tasks} = prepareLinkShare()

		cy.visit(`/tasks/${tasks[0].id}#share-auth-token=${share.hash}`)

		cy.get('h1.title')
			.should('contain', tasks[0].title)
	})
})
