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

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should be an empty project', () => {
		cy.visit('/projects/1')
		cy.url()
			.should('contain', '/projects/1/1')
		cy.get('.project-title')
			.should('contain', 'First Project')
		cy.get('.project-title-dropdown')
			.should('exist')
		cy.get('p')
			.contains('This project is currently empty.')
			.should('exist')
	})
	
	// FIXME: Migrated to Playwright - skip to avoid duplication

	
	it.skip('Should create a new task', () => {
		BucketFactory.create(2, {
			project_view_id: 4,
		})
		
		const newTaskTitle = 'New task'
		
		cy.visit('/projects/1')
		cy.get('.task-add textarea')
			.type(newTaskTitle+'{enter}')
		cy.get('.tasks')
			.should('contain.text', newTaskTitle)
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should navigate to the task when the title is clicked', () => {
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		cy.visit('/projects/1/1')

		cy.get('.tasks .task .tasktext')
			.contains(tasks[0].title)
			.first()
			.click()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`)
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should not see any elements for a project which is shared read only', () => {
		UserFactory.create(2)
		UserProjectFactory.create(1, {
			project_id: 2,
			user_id: 1,
			permission: 0,
		})
		const projects = ProjectFactory.create(2, {
			owner_id: '{increment}',
		})
		cy.visit(`/projects/${projects[1].id}/`)

		cy.get('.project-title-wrapper .icon')
			.should('not.exist')
		cy.get('input.input[placeholder="Add a task…"]')
			.should('not.exist')
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should only show the color of a project in the navigation and not in the list view', () => {
		const projects = ProjectFactory.create(1, {
			hex_color: '00db60',
		})
		TaskFactory.create(10, {
			project_id: projects[0].id,
		})
		cy.visit(`/projects/${projects[0].id}/`)

		cy.get('.menu-list li .list-menu-link .color-bubble')
			.should('have.css', 'background-color', 'rgb(0, 219, 96)')
		cy.get('.tasks .color-bubble')
			.should('not.exist')
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should paginate for > 50 tasks', () => {
		const tasks = TaskFactory.create(100, {
			id: '{increment}',
			title: i => `task${i}`,
			project_id: 1,
		})
		cy.visit('/projects/1/1')

		cy.get('.tasks')
			.should('contain', tasks[20].title)
		cy.get('.tasks')
			.should('not.contain', tasks[99].title)

		cy.get('.card-content .pagination .pagination-link')
			.contains('2')
			.click()

		cy.url()
			.should('contain', '?page=2')
		cy.get('.tasks')
			.should('contain', tasks[99].title)
		cy.get('.tasks')
			.should('not.contain', tasks[20].title)
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should show cross-project subtasks in their own project List view', () => {
		const projects = createProjects(2)

		const tasks = [
			TaskFactory.create(1, {
				id: 1,
				title: 'Parent Task in Project A',
				project_id: projects[0].id,
			}, false)[0],
			TaskFactory.create(1, {
				id: 2,
				title: 'Subtask in Project B',
				project_id: projects[1].id,
			}, false)[0],
		]

		// Make task 2 a subtask of task 1
		TaskRelationFactory.truncate()
		TaskRelationFactory.create(1, {
			id: 1,
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		}, false)
		TaskRelationFactory.create(1, {
			id: 2,
			task_id: 1,
			other_task_id: 2,
			relation_kind: 'parenttask',
		}, false)

		cy.visit(`/projects/${projects[1].id}/${projects[1].views[0].id}`)

		cy.get('.tasks')
			.should('contain', 'Subtask in Project B')
	})

	// FIXME: Migrated to Playwright - skip to avoid duplication


	it.skip('Should show same-project subtasks under their parent', () => {
		const projects = createProjects(1)

		const tasks = [
			TaskFactory.create(1, {
				id: 1,
				title: 'Parent Task',
				project_id: projects[0].id,
			}, false)[0],
			TaskFactory.create(1, {
				id: 2,
				title: 'Subtask Same Project',
				project_id: projects[0].id,
			}, false)[0],
		]

		// Make task 2 a subtask of task 1
		TaskRelationFactory.truncate()
		TaskRelationFactory.create(1, {
			id: 1,
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		}, false)
		TaskRelationFactory.create(1, {
			id: 2,
			task_id: 1,
			other_task_id: 2,
			relation_kind: 'parenttask',
		}, false)

		cy.visit(`/projects/${projects[0].id}/${projects[0].views[0].id}`)

		cy.get('.tasks')
			.should('contain', 'Parent Task')
		cy.get('.tasks')
			.should('contain', 'Subtask Same Project')

		cy.get('ul.tasks > div > .single-task')
			.should('exist')
		cy.get('ul.tasks > div > .subtask-nested')
			.should('exist')
	})

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
