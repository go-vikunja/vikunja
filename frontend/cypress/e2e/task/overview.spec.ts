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
})
