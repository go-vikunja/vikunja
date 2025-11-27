import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskRelationFactory} from '../../factories/task_relation'

async function createViews(projectId: number, projectViewId: number) {
	return (await ProjectViewFactory.create(1, {
		id: projectViewId,
		project_id: projectId,
		view_kind: 0,
	}, false))[0]
}

test.describe('Subtask duplicate handling', () => {
	let projectA
	let projectB
	let parentA
	let parentB
	let subtask

	test.beforeEach(async ({authenticatedPage: page, apiContext}) => {
		await Promise.all([
			ProjectFactory.truncate(),
			ProjectViewFactory.truncate(),
			TaskFactory.truncate(),
			TaskRelationFactory.truncate(),
		])

		projectA = (await ProjectFactory.create(1, {id: 1, title: 'Project A'}))[0]
		await createViews(projectA.id, 1)
		projectB = (await ProjectFactory.create(1, {id: 2, title: 'Project B'}, false))[0]
		await createViews(projectB.id, 2)

		parentA = (await TaskFactory.create(1, {id: 10, title: 'Parent A', project_id: projectA.id}, false))[0]
		parentB = (await TaskFactory.create(1, {id: 11, title: 'Parent B', project_id: projectB.id}, false))[0]
		subtask = (await TaskFactory.create(1, {id: 12, title: 'Shared subtask', project_id: projectA.id}, false))[0]

		// Navigate to a page first to establish context for localStorage access
		await page.goto('/')
		const token = await page.evaluate(() => localStorage.getItem('token'))

		await apiContext.put(`tasks/${parentA.id}/relations`, {
			headers: {
				'Authorization': `Bearer ${token}`,
			},
			data: {
				other_task_id: subtask.id,
				relation_kind: 'subtask',
			},
		})

		await apiContext.put(`tasks/${parentB.id}/relations`, {
			headers: {
				'Authorization': `Bearer ${token}`,
			},
			data: {
				other_task_id: subtask.id,
				relation_kind: 'subtask',
			},
		})
	})

	test('shows subtask only once in each project list', async ({authenticatedPage: page}) => {
		await page.goto(`/projects/${projectA.id}/1`)
		await expect(page.locator('.subtask-nested .task-link').filter({hasText: subtask.title})).toBeVisible()
		await expect(page.locator('.tasks .task-link').filter({hasText: subtask.title})).toHaveCount(1)

		await page.goto(`/projects/${projectB.id}/1`)
		await expect(page.locator('.subtask-nested .task-link').filter({hasText: subtask.title})).toBeVisible()
		await expect(page.locator('.tasks .task-link').filter({hasText: subtask.title})).toHaveCount(1)
	})
})
