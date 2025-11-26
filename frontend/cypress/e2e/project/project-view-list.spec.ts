import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {UserProjectFactory} from '../../factories/users_project'
import {TaskFactory} from '../../factories/task'
import {TaskRelationFactory} from '../../factories/task_relation'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {prepareProjects, createProjects} from './prepareProjects'
import {BucketFactory} from '../../factories/bucket'
import {
	createTasksWithPriorities,
	createTasksWithSearch,
} from '../../support/filterTestHelpers'

describe('Project View List', () => {
	createFakeUserAndLogin()
	prepareProjects()

	it('Should respect filter query parameter from URL', () => {
		const {highPriorityTasks, lowPriorityTasks} = createTasksWithPriorities()

		cy.visit('/projects/1/1?filter=priority%20>=%204')

		cy.url()
			.should('include', 'filter=priority')

		cy.contains('.tasks', highPriorityTasks[0].title, {timeout: 10000})
			.should('exist')

		cy.get('.tasks')
			.should('contain', highPriorityTasks[0].title)
		cy.get('.tasks')
			.should('contain', highPriorityTasks[1].title)

		cy.get('.tasks')
			.should('not.contain', lowPriorityTasks[0].title)
		cy.get('.tasks')
			.should('not.contain', lowPriorityTasks[1].title)
	})

	it('Should respect search query parameter from URL', () => {
		const {searchableTask} = createTasksWithSearch()

		cy.visit('/projects/1/1?s=meeting')

		cy.url()
			.should('include', 's=meeting')

		cy.contains('.tasks', searchableTask.title, {timeout: 10000})
			.should('exist')

		cy.get('.tasks')
			.should('contain', searchableTask.title)

		cy.get('.tasks .task')
			.should('have.length', 1)
	})
})
