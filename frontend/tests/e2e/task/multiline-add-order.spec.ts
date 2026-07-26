import type {Page} from '@playwright/test'

import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from '../project/prepareProjects'

const MULTILINE_INPUT = [
	'Parent Task',
	'    Subtask 1',
	'    Subtask 2',
	'    Subtask 3',
	'    Subtask 4',
	'    Subtask 5',
	'    Subtask 6',
].join('\n')

const EXPECTED_ORDER = [
	'Parent Task',
	'Subtask 1',
	'Subtask 2',
	'Subtask 3',
	'Subtask 4',
	'Subtask 5',
	'Subtask 6',
]

async function addMultilineTasks(page: Page) {
	const textarea = page.locator('.task-add textarea')
	await textarea.fill(MULTILINE_INPUT)
	await textarea.press('Enter')

	await expect(page.locator('.tasks .tasktext').filter({hasText: 'Subtask 6'})).toBeVisible()
}

test.describe('Multi-line quick add ordering', () => {
	test('keeps the entered order in an empty project', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await page.goto('/projects/1/1')

		await addMultilineTasks(page)
		await page.reload()

		await expect(page.locator('.tasks .tasktext')).toHaveText(EXPECTED_ORDER)
	})

	test('keeps the entered order in a project with existing tasks', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: 1,
		})
		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks .task')).toHaveCount(3)

		await addMultilineTasks(page)
		await page.reload()
		await expect(page.locator('.tasks .tasktext')).toHaveCount(EXPECTED_ORDER.length + 3)

		const titles = await page.locator('.tasks .tasktext').allInnerTexts()
		expect(titles.slice(0, EXPECTED_ORDER.length)).toEqual(EXPECTED_ORDER)
	})
})
