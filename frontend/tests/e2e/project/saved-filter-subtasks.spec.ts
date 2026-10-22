import type {APIRequestContext} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {TaskFactory} from '../../factories/task'

async function setupFilterWithSubtask(apiContext: APIRequestContext, userToken: string, parentPriority: number, subtaskPriority: number) {
	const project = (await ProjectFactory.create(1, {id: 1, title: 'Project'}))[0]
	await ProjectViewFactory.create(1, {
		id: 1,
		project_id: project.id,
		view_kind: 0,
	}, false)

	const parent = (await TaskFactory.create(1, {
		id: 10,
		title: 'Parent task',
		project_id: project.id,
		priority: parentPriority,
	}, false))[0]
	const subtask = (await TaskFactory.create(1, {
		id: 11,
		title: 'Subtask',
		project_id: project.id,
		priority: subtaskPriority,
	}, false))[0]

	const headers = {'Authorization': `Bearer ${userToken}`}

	await apiContext.put(`tasks/${parent.id}/relations`, {
		headers,
		data: {
			other_task_id: subtask.id,
			relation_kind: 'subtask',
		},
	})

	const filterResponse = await apiContext.put('filters', {
		headers,
		data: {
			title: 'Priority filter',
			filters: {
				filter: 'done = false && priority = 5',
				filter_include_nulls: false,
				s: '',
			},
		},
	})
	const filter = await filterResponse.json()

	return {parent, subtask, filterProjectId: filter.id * -1 - 1}
}

test.describe('Saved filter subtasks', () => {
	test('shows a non-matching subtask nested only, not as its own top level row', async ({authenticatedPage: page, apiContext, userToken}) => {
		const {parent, subtask, filterProjectId} = await setupFilterWithSubtask(apiContext, userToken, 5, 1)

		await page.goto(`/projects/${filterProjectId}`)

		await expect(page.locator('.tasks .task-link').filter({hasText: parent.title})).toHaveCount(1)
		await expect(page.locator('.subtask-nested .task-link').filter({hasText: subtask.title})).toBeVisible()
		await expect(page.locator('.tasks .task-link').filter({hasText: subtask.title})).toHaveCount(1)
	})

	// #2494: the matching subtask is the only result, its parent must not hide it
	test('shows a matching subtask when its parent does not match the filter', async ({authenticatedPage: page, apiContext, userToken}) => {
		const {parent, subtask, filterProjectId} = await setupFilterWithSubtask(apiContext, userToken, 1, 5)

		await page.goto(`/projects/${filterProjectId}`)

		await expect(page.locator('.tasks .task-link').filter({hasText: subtask.title})).toHaveCount(1)
		await expect(page.locator('.tasks .task-link').filter({hasText: parent.title})).toHaveCount(0)
	})
})
