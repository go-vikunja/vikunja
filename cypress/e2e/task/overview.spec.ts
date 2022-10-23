import {ListFactory} from '../../factories/list'
import {seed} from '../../support/seed'
import {TaskFactory} from '../../factories/task'
import {formatISO} from 'date-fns'
import {UserFactory} from '../../factories/user'
import {NamespaceFactory} from '../../factories/namespace'
import {BucketFactory} from '../../factories/bucket'
import {updateUserSettings} from '../../support/updateUserSettings'

import '../../support/authenticateUser'

function seedTasks(numberOfTasks = 100, startDueDate = new Date()) {
	UserFactory.create(1)
	NamespaceFactory.create(1)
	const list = ListFactory.create()[0]
	BucketFactory.create(1, {
		list_id: list.id,
	})
	const tasks = []
	let dueDate = startDueDate
	for (let i = 0; i < numberOfTasks; i++) {
		const now = new Date()
		dueDate = (new Date(dueDate.valueOf())).setDate((new Date(dueDate.valueOf())).getDate() + 2)
		tasks.push({
			id: i + 1,
			list_id: list.id,
			done: false,
			created_by_id: 1,
			title: 'Test Task ' + i,
			index: i + 1,
			due_date: formatISO(dueDate),
			created: formatISO(now),
			updated: formatISO(now),
		})
	}
	seed(TaskFactory.table, tasks)
	return {tasks, list}
}

describe('Home Page Task Overview', () => {
	it('Should show tasks with a near due date first on the home page overview', () => {
		const {tasks} = seedTasks()

		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.each(([task], index) => {
				expect(task.innerText).to.contain(tasks[index].title)
			})
	})

	it('Should show overdue tasks first, then show other tasks', () => {
		const oldDate = (new Date()).setDate((new Date()).getDate() - 14)
		const {tasks} = seedTasks(100, oldDate)

		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.each(([task], index) => {
				expect(task.innerText).to.contain(tasks[index].title)
			})
	})

	it('Should show a new task with a very soon due date at the top', () => {
		const {tasks} = seedTasks()
		const newTaskTitle = 'New Task'
		
		cy.visit('/')
		
		TaskFactory.create(1, {
			id: 999,
			title: newTaskTitle,
			due_date: formatISO(new Date()),
		}, false)
		
		cy.visit(`/lists/${tasks[0].list_id}/list`)
		cy.get('.tasks .task')
			.first()
			.should('contain.text', newTaskTitle)
		cy.visit('/')
		cy.get('[data-cy="showTasks"] .card .task')
			.first()
			.should('contain.text', newTaskTitle)
	})
	
	it('Should not show a new task without a date at the bottom when there are > 50 tasks', () => {
		// We're not using the api here to create the task in order to verify the flow
		const {tasks} = seedTasks()
		const newTaskTitle = 'New Task'

		cy.visit('/')

		cy.visit(`/lists/${tasks[0].list_id}/list`)
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
	
	it('Should show a task without a due date added via default list at the bottom', () => {
		const {list} = seedTasks(40)
		updateUserSettings({
			default_list_id: list.id,
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
	
	it('Should show the cta buttons for new list when there are no tasks', () => {
		TaskFactory.truncate()
		
		cy.visit('/')
		
		cy.get('.home.app-content .content')
			.should('contain.text', 'You can create a new list for your new tasks:')
			.should('contain.text', 'Or import your lists and tasks from other services into Vikunja:')
	})
	
	it('Should not show the cta buttons for new list when there are tasks', () => {
		seedTasks()

		cy.visit('/')

		cy.get('.home.app-content .content')
			.should('not.contain.text', 'You can create a new list for your new tasks:')
			.should('not.contain.text', 'Or import your lists and tasks from other services into Vikunja:')
	})
})
