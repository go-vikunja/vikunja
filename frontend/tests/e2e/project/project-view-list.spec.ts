import {test, expect} from '../../support/fixtures'
import {UserProjectFactory} from '../../factories/users_project'
import {TaskFactory} from '../../factories/task'
import {TaskRelationFactory} from '../../factories/task_relation'
import {UserFactory} from '../../factories/user'
import {ProjectFactory} from '../../factories/project'
import {createProjects} from './prepareProjects'
import {BucketFactory} from '../../factories/bucket'

test.describe('Project View List', () => {
	test('Should be an empty project', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await page.goto('/projects/1')
		await expect(page).toHaveURL(/\/projects\/1\/1/)
		await expect(page.locator('.project-title')).toContainText('First Project')
		await expect(page.locator('.project-title-dropdown')).toBeVisible()
		await expect(page.locator('.has-text-centered.has-text-grey.is-italic').filter({hasText: 'This project is currently empty.'})).toBeVisible()
	})

	test('Should create a new task', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await BucketFactory.create(2, {
			project_view_id: 4,
		})

		const newTaskTitle = 'New task'

		await page.goto('/projects/1/1')
		await page.locator('.task-add textarea').fill(newTaskTitle)
		await page.locator('.task-add textarea').press('Enter')
		await expect(page.locator('.tasks')).toContainText(newTaskTitle)
	})

	test('Should navigate to the task when the title is clicked', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const tasks = await TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})
		await page.goto('/projects/1/1')

		await page.locator('.tasks .task .tasktext').filter({hasText: tasks[0].title}).first().click()

		await expect(page).toHaveURL(new RegExp(`/tasks/${tasks[0].id}`))
	})

	test('Should not see any elements for a project which is shared read only', async ({authenticatedPage: page}) => {
		await UserFactory.create(2)
		await UserProjectFactory.create(1, {
			project_id: 2,
			user_id: 1,
			permission: 0,
		})
		const projects = await ProjectFactory.create(2, {
			owner_id: '{increment}',
		})
		await page.goto(`/projects/${projects[1].id}/`)

		await expect(page.locator('.project-title-wrapper .icon')).not.toBeVisible()
		await expect(page.locator('input.input[placeholder="Add a task…"]')).not.toBeVisible()
	})

	test('Should only show the color of a project in the navigation and not in the list view', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1, {
			id: 1,
			hex_color: '00db60',
		})
		await TaskFactory.create(10, {
			project_id: projects[0].id,
		})
		await page.goto(`/projects/${projects[0].id}/1`)

		await expect(page.locator('.menu-list li .list-menu-link .color-bubble')).toHaveCSS('background-color', 'rgb(0, 219, 96)')
		await expect(page.locator('.tasks .color-bubble')).not.toBeVisible()
	})

	test('Should paginate for > 50 tasks', async ({authenticatedPage: page}) => {
		await createProjects(1)
		const tasks = await TaskFactory.create(100, {
			id: '{increment}',
			title: i => `task${i}`,
			project_id: 1,
		})
		await page.goto('/projects/1/1')

		await expect(page.locator('.tasks')).toContainText(tasks[20].title)
		await expect(page.locator('.tasks')).not.toContainText(tasks[99].title)

		await page.locator('.card-content .pagination .pagination-link').filter({hasText: '2'}).click()

		await expect(page).toHaveURL(/\?page=2/)
		await expect(page.locator('.tasks')).toContainText(tasks[99].title)
		await expect(page.locator('.tasks')).not.toContainText(tasks[20].title)
	})

	test('Should show cross-project subtasks in their own project List view', async ({authenticatedPage: page}) => {
		const projects = await createProjects(2)

		await TaskFactory.create(1, {
			id: 1,
			title: 'Parent Task in Project A',
			project_id: projects[0].id,
		}, false)
		await TaskFactory.create(1, {
			id: 2,
			title: 'Subtask in Project B',
			project_id: projects[1].id,
		}, false)

		// Make task 2 a subtask of task 1
		TaskRelationFactory.truncate()
		await TaskRelationFactory.create(1, {
			id: 1,
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		}, false)
		await TaskRelationFactory.create(1, {
			id: 2,
			task_id: 1,
			other_task_id: 2,
			relation_kind: 'parenttask',
		}, false)

		await page.goto(`/projects/${projects[1].id}/${projects[1].views[0].id}`)

		await expect(page.locator('.tasks')).toContainText('Subtask in Project B')
	})

	test('Should show same-project subtasks under their parent', async ({authenticatedPage: page}) => {
		const projects = await createProjects(1)

		await TaskFactory.create(1, {
			id: 1,
			title: 'Parent Task',
			project_id: projects[0].id,
		}, false)
		await TaskFactory.create(1, {
			id: 2,
			title: 'Subtask Same Project',
			project_id: projects[0].id,
		}, false)

		// Make task 2 a subtask of task 1
		TaskRelationFactory.truncate()
		await TaskRelationFactory.create(1, {
			id: 1,
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		}, false)
		await TaskRelationFactory.create(1, {
			id: 2,
			task_id: 1,
			other_task_id: 2,
			relation_kind: 'parenttask',
		}, false)

		await page.goto(`/projects/${projects[0].id}/${projects[0].views[0].id}`)

		await expect(page.locator('.tasks')).toContainText('Parent Task')
		await expect(page.locator('.tasks')).toContainText('Subtask Same Project')

		await expect(page.locator('ul.tasks > div > .single-task')).toBeVisible()
		await expect(page.locator('ul.tasks > div > .subtask-nested')).toBeVisible()
	})
})
