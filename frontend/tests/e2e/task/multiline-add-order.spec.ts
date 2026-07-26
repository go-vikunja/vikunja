import type {Page} from '@playwright/test'

import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {TaskPositionFactory} from '../../factories/task_position'
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
		// The existing tasks need real positions: without them the api falls back to
		// deriving one from the task index, which hides the ordering bug.
		await TaskPositionFactory.create(3, {
			task_id: '{increment}',
			project_view_id: 1,
		})
		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks .task')).toHaveCount(3)

		await addMultilineTasks(page)
		await page.reload()
		await expect(page.locator('.tasks .tasktext')).toHaveCount(EXPECTED_ORDER.length + 3)

		const titles = await page.locator('.tasks .tasktext').allInnerTexts()
		expect(titles.slice(0, EXPECTED_ORDER.length)).toEqual(EXPECTED_ORDER)
	})

	// Unlike subtasks, which render nested in relation order, flat tasks are list rows
	// sorted by position - so this is the case that needs a position per task.
	test('keeps the entered order for a flat list of tasks', async ({authenticatedPage: page}) => {
		await createProjects(1)
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: 1,
		})
		await TaskPositionFactory.create(3, {
			task_id: '{increment}',
			project_view_id: 1,
		})
		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks .task')).toHaveCount(3)

		const flat = ['Flat 1', 'Flat 2', 'Flat 3', 'Flat 4', 'Flat 5', 'Flat 6', 'Flat 7']
		const textarea = page.locator('.task-add textarea')
		await textarea.fill(flat.join('\n'))
		await textarea.press('Enter')
		await expect(page.locator('.tasks .tasktext').filter({hasText: flat.at(-1)})).toBeVisible()

		// guards the pre-reload insertion order, not just the order after reload
		const before = await page.locator('.tasks .tasktext').allInnerTexts()
		expect(before.slice(0, flat.length)).toEqual(flat)

		await page.reload()
		await expect(page.locator('.tasks .tasktext')).toHaveCount(flat.length + 3)

		const titles = await page.locator('.tasks .tasktext').allInnerTexts()
		expect(titles.slice(0, flat.length)).toEqual(flat)
	})

	// A batch big enough that halving the lowest position for every task would run out
	// of room and trigger a full position recalculation part way through, which drops
	// tasks out of sequence. Guards against replacing the computed positions with
	// "create back to front and let the api put each one on top".
	test('keeps the entered order for a batch large enough to exhaust the gap', async ({authenticatedPage: page}) => {
		test.setTimeout(120000)
		await createProjects(1)
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: 1,
		})
		await TaskPositionFactory.create(3, {
			task_id: '{increment}',
			project_view_id: 1,
		})
		await page.goto('/projects/1/1')
		await expect(page.locator('.tasks .task')).toHaveCount(3)

		const lines = Array.from({length: 60}, (_, i) => `Big ${String(i + 1).padStart(2, '0')}`)
		const textarea = page.locator('.task-add textarea')
		await textarea.fill(lines.join('\n'))
		await textarea.press('Enter')
		await expect(page.locator('.tasks .tasktext').filter({hasText: lines.at(-1)})).toBeVisible({timeout: 60000})

		await page.reload()
		// The new tasks all sort above the existing three, so page one is exactly the
		// first 50 of them.
		await expect(page.locator('.tasks .tasktext')).toHaveCount(50)
		expect(await page.locator('.tasks .tasktext').allInnerTexts()).toEqual(lines.slice(0, 50))
	})
})
