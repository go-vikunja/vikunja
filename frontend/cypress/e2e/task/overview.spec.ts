import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {ProjectFactory} from '../../factories/project'
import {seed} from '../../support/seed'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {updateUserSettings} from '../../support/updateUserSettings'
import {createDefaultViews} from "../project/prepareProjects";

function seedTasks(numberOfTasks = 50, startDueDate = new Date()) {
	const project = ProjectFactory.create()[0]
	const views = createDefaultViews(project.id)
	BucketFactory.create(1, {
		project_view_id: views[3].id,
	})
	const tasks = []
	let dueDate = startDueDate
	for (let i = 0; i < numberOfTasks; i++) {
		const now = new Date()
		dueDate = new Date(new Date(dueDate).setDate(dueDate.getDate() + 2))
		tasks.push({
			id: i + 1,
			project_id: project.id,
			done: false,
			created_by_id: 1,
			title: 'Test Task ' + i,
			index: i + 1,
			due_date: dueDate.toISOString(),
			created: now.toISOString(),
			updated: now.toISOString(),
		})
	}
	seed(TaskFactory.table, tasks)
	return {tasks, project}
}

describe('Home Page Task Overview', () => {
	createFakeUserAndLogin()

	it('Should show tasks with a near due date first on the home page overview', () => {
		const taskCount = 50
		const {tasks} = seedTasks(taskCount)

		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.each(([task], index) => {
				expect(task.innerText).to.contain(tasks[index].title)
			})
	})

	it('Should show overdue tasks first, then show other tasks', () => {
		const now = new Date()
		const oldDate = new Date(new Date(now).setDate(now.getDate() - 14))
		const taskCount = 50
		const {tasks} = seedTasks(taskCount, oldDate)

		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.each(([task], index) => {
				expect(task.innerText).to.contain(tasks[index].title)
			})
	})

	it('Should show a new task with a very soon due date at the top', () => {
		const {tasks} = seedTasks(49)
		const newTaskTitle = 'New Task'
		
		cy.visit('/')
		
		TaskFactory.create(1, {
			id: 999,
			title: newTaskTitle,
			due_date: new Date().toISOString(),
		}, false)
		
		cy.visit(`/projects/${tasks[0].project_id}/1`)
		cy.get('.tasks .task')
			.should('contain.text', newTaskTitle)
		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', newTaskTitle)
	})
	
	it('Should not show a new task without a date at the bottom when there are > 50 tasks', () => {
		// We're not using the api here to create the task in order to verify the flow
		const {tasks} = seedTasks(100)
		const newTaskTitle = 'New Task'

		cy.visit('/')

		cy.visit(`/projects/${tasks[0].project_id}/1`)
		cy.get('.task-add textarea')
			.type(newTaskTitle+'{enter}')
		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.last()
			.should('not.contain.text', newTaskTitle)
	})
	
	it('Should show a new task without a date at the bottom when there are < 50 tasks', () => {
		seedTasks(40)
		const newTaskTitle = 'New Task'
		TaskFactory.create(1, {
			id: 999,
			title: newTaskTitle,
		}, false)

		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.last()
			.should('contain.text', newTaskTitle)
	})
	
	it('Should show a task without a due date added via default project at the bottom', () => {
		const {project} = seedTasks(40)
		updateUserSettings({
			default_project_id: project.id,
			overdue_tasks_reminders_time: '9:00',
		})
		
		const newTaskTitle = 'New Task'
		cy.visit('/')
		
		cy.get('.add-task-textarea')
			.type(`${newTaskTitle}{enter}`)
		
		cy.get('[data-cy="showTasks"] .card .task')
			.last()
			.should('contain.text', newTaskTitle)
	})
	
	it('Should show the cta buttons for new project when there are no tasks', () => {
		TaskFactory.truncate()
		
		cy.visit('/')
		
		cy.get('.home.app-content .content')
			.should('contain.text', 'Import your projects and tasks from other services into Vikunja:')
	})
	
	it('Should not show the cta buttons for new project when there are tasks', () => {
		seedTasks()

		cy.visit('/')

		cy.get('.home.app-content .content')
			.should('not.contain.text', 'You can create a new project for your new tasks:')
			.should('not.contain.text', 'Or import your projects and tasks from other services into Vikunja:')
	})

	it('Should show sort controls on the overview page', () => {
		seedTasks()

		cy.visit('/')

		cy.get('.title-container .sort-container')
			.should('exist')
			.should('be.visible')
		cy.get('.title-container .sort-container button')
			.should('have.length', 2)
	})

	it('Should toggle between due date and priority sorting', () => {
		const now = new Date()
		const project = ProjectFactory.create()[0]
		const views = createDefaultViews(project.id)
		BucketFactory.create(1, {
			project_view_id: views[3].id,
		})

		// Create tasks with different priorities and due dates
		const tasks = [
			{
				id: 1,
				project_id: project.id,
				title: 'Low Priority Soon',
				priority: 1,
				due_date: new Date(now.getTime() + 86400000).toISOString(), // tomorrow
				created: now.toISOString(),
				updated: now.toISOString(),
			},
			{
				id: 2,
				project_id: project.id,
				title: 'High Priority Later',
				priority: 5,
				due_date: new Date(now.getTime() + 172800000).toISOString(), // 2 days
				created: now.toISOString(),
				updated: now.toISOString(),
			},
		]
		seed(TaskFactory.table, tasks)

		cy.visit('/')

		// Default: sorted by due date (ascending)
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', 'Low Priority Soon')

		// Click to sort by priority
		cy.get('.title-container .sort-container button')
			.first()
			.click()

		// Should now be sorted by priority (ascending, so low priority first)
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', 'Low Priority Soon')
	})

	it('Should toggle between ascending and descending order', () => {
		const taskCount = 3
		const {tasks} = seedTasks(taskCount)

		cy.visit('/')

		// Default: ascending order (earliest due date first)
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', tasks[0].title)

		// Click to toggle to descending
		cy.get('.title-container .sort-container button')
			.last()
			.click()

		// Should now be in descending order (latest due date first)
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', tasks[taskCount - 1].title)

		// Click again to toggle back to ascending
		cy.get('.title-container .sort-container button')
			.last()
			.click()

		// Should be back to ascending order
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', tasks[0].title)
	})

	it('Should persist sort preferences in URL', () => {
		seedTasks()

		cy.visit('/')

		// Change to priority sorting
		cy.get('.title-container .sort-container button')
			.first()
			.click()

		cy.url().should('include', 'sortBy=priority')

		// Change to descending order
		cy.get('.title-container .sort-container button')
			.last()
			.click()

		cy.url().should('include', 'sortOrder=desc')

		// Refresh page and verify settings persist
		cy.reload()

		cy.url().should('include', 'sortBy=priority')
		cy.url().should('include', 'sortOrder=desc')
	})
})
