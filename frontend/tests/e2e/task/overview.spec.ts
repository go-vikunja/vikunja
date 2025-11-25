import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {seed} from '../../support/seed'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {updateUserSettings} from '../../support/updateUserSettings'
import {createDefaultViews} from '../project/prepareProjects'
import type {APIRequestContext} from '@playwright/test'

async function seedTasks(apiContext: APIRequestContext, numberOfTasks = 50, startDueDate = new Date()) {
	const project = (await ProjectFactory.create())[0]
	const views = await createDefaultViews(project.id)
	await BucketFactory.create(1, {
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
	await TaskFactory.seed(TaskFactory.table, tasks)
	return {tasks, project}
}

test.describe('Home Page Task Overview', () => {
	test('Should show tasks with a near due date first on the home page overview', async ({authenticatedPage: page, apiContext}) => {
		const taskCount = 50
		const {tasks} = await seedTasks(apiContext, taskCount)

		await page.goto('/')
		const taskElements = await page.locator('[data-cy="showTasks"] .card .task').all()
		for (let index = 0; index < taskElements.length; index++) {
			const taskText = await taskElements[index].innerText()
			expect(taskText).toContain(tasks[index].title)
		}
	})

	test('Should show overdue tasks first, then show other tasks', async ({authenticatedPage: page, apiContext}) => {
		const now = new Date()
		const oldDate = new Date(new Date(now).setDate(now.getDate() - 14))
		const taskCount = 50
		const {tasks} = await seedTasks(apiContext, taskCount, oldDate)

		await page.goto('/')
		const taskElements = await page.locator('[data-cy="showTasks"] .card .task').all()
		for (let index = 0; index < taskElements.length; index++) {
			const taskText = await taskElements[index].innerText()
			expect(taskText).toContain(tasks[index].title)
		}
	})

	// FIXME: Selector '.tasks .task' resolves to 50 elements causing strict mode violation
	test.skip('Should show a new task with a very soon due date at the top', async ({authenticatedPage: page, apiContext}) => {
		const {tasks} = await seedTasks(apiContext, 49)
		const newTaskTitle = 'New Task'

		await page.goto('/')

		await TaskFactory.create(1, {
			id: 999,
			title: newTaskTitle,
			due_date: new Date().toISOString(),
		}, false)

		await page.goto(`/projects/${tasks[0].project_id}/1`)
		await expect(page.locator('.tasks .task')).toContainText(newTaskTitle)
		await page.goto('/')
		await expect(page.locator('[data-cy="showTasks"] .card .task').first()).toContainText(newTaskTitle)
	})

	test('Should not show a new task without a date at the bottom when there are > 50 tasks', async ({authenticatedPage: page, apiContext}) => {
		// We're not using the api here to create the task in order to verify the flow
		const {tasks} = await seedTasks(apiContext, 100)
		const newTaskTitle = 'New Task'

		await page.goto('/')

		await page.goto(`/projects/${tasks[0].project_id}/1`)
 		const taskResponsePromise = page.waitForResponse('**/api/v1/projects/*/tasks');
		await page.locator('.task-add textarea').fill(newTaskTitle)
		await page.locator('.task-add textarea').press('Enter')
		await taskResponsePromise
		await page.goto('/')
		await expect(page.locator('[data-cy="showTasks"]').last()).not.toContainText(newTaskTitle)
	})

	// FIXME: the task is not shown 
	test.skip('Should show a new task without a date at the bottom when there are < 50 tasks', async ({authenticatedPage: page, apiContext}) => {
		await seedTasks(apiContext, 40)
		const newTaskTitle = 'New Task'
		await TaskFactory.create(1, {
			id: 999,
			title: newTaskTitle,
		}, false)

		await page.goto('/')
		await expect(page.locator('[data-cy="showTasks"]')).toContainText(newTaskTitle)
	})

	// FIXME: SecurityError when accessing localStorage - "Access is denied for this document"
	test.skip('Should show a task without a due date added via default project at the bottom', async ({authenticatedPage: page, apiContext}) => {
		const {project} = await seedTasks(apiContext, 40)
		const token = await page.evaluate(() => localStorage.getItem('token'))
		await updateUserSettings(apiContext, token, {
			default_project_id: project.id,
			overdue_tasks_reminders_time: '9:00',
		})

		const newTaskTitle = 'New Task'
		await page.goto('/')

		await page.locator('.add-task-textarea').fill(newTaskTitle)
		await page.locator('.add-task-textarea').press('Enter')

		await expect(page.locator('[data-cy="showTasks"] .card .task').last()).toContainText(newTaskTitle)
	})

	test('Should show the cta buttons for new project when there are no tasks', async ({authenticatedPage: page}) => {
		TaskFactory.truncate()

		await page.goto('/')

		await expect(page.locator('.home.app-content .content')).toContainText('Import your projects and tasks from other services into Vikunja:')
	})

	test('Should not show the cta buttons for new project when there are tasks', async ({authenticatedPage: page, apiContext}) => {
		await seedTasks(apiContext)

		await page.goto('/')

		await expect(page.locator('.home.app-content .content')).not.toContainText('You can create a new project for your new tasks:')
		await expect(page.locator('.home.app-content .content')).not.toContainText('Or import your projects and tasks from other services into Vikunja:')
	})
})
