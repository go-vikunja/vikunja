import {test, expect} from '../../support/fixtures'
import {clickTaskAction} from '../../support/taskDetail'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

const BASE_TIME = Date.parse('2026-01-01T00:00:00Z')

function minutesAfterBase(minutes: number) {
	return new Date(BASE_TIME + minutes * 60_000).toISOString()
}

async function seedProjectsWithTasks() {
	const currentProject = (await ProjectFactory.create(1, {id: 1, title: 'Current Project'}))[0]
	await createDefaultViews(currentProject.id)
	const otherProject = (await ProjectFactory.create(1, {id: 2, title: 'Other Project'}, false))[0]
	await createDefaultViews(otherProject.id, 5)

	// The parent is the most recently updated task, so excluding it must not shrink the list.
	const parent = (await TaskFactory.create(1, {
		id: 1,
		title: 'Parent task',
		project_id: currentProject.id,
		updated: minutesAfterBase(1000),
	}, false))[0]
	await TaskFactory.create(12, {
		id: (i: number) => 100 + i,
		index: (i: number) => 100 + i,
		title: (i: number) => `Suggestion ${i}`,
		project_id: currentProject.id,
		done: (i: number) => i === 10,
		updated: (i: number) => minutesAfterBase(i),
	}, false)
	await TaskFactory.create(1, {
		id: 200,
		title: 'Recent task elsewhere',
		project_id: otherProject.id,
		updated: minutesAfterBase(900),
	}, false)

	return parent
}

async function openRelatedTasksForm(page) {
	await clickTaskAction(page, 'Add relation')
	const input = page.locator('.task-relations .multiselect input').first()
	await expect(input).toBeVisible()
	return input
}

function relationTaskSearchResults(page) {
	return page.locator('.task-relations .multiselect .search-results .search-result-button:not(.is-create-option)')
}

test.describe('Related tasks suggestions', () => {
	test('Shows the 10 most recently updated tasks of the current project before typing', async ({authenticatedPage: page}) => {
		const parent = await seedProjectsWithTasks()

		await page.goto(`/tasks/${parent.id}`)
		await openRelatedTasksForm(page)

		const results = relationTaskSearchResults(page)
		await expect(results).toHaveCount(10, {timeout: 10000})
		// Done tasks keep their place: the order is purely by last update.
		const expectedTitles = [12, 11, 10, 9, 8, 7, 6, 5, 4, 3].map(i => `Suggestion ${i}`)
		for (const [index, title] of expectedTitles.entries()) {
			await expect(results.nth(index)).toContainText(title)
		}
		await expect(results.filter({hasText: 'Parent task'})).toHaveCount(0)
		await expect(results.filter({hasText: 'Recent task elsewhere'})).toHaveCount(0)
	})

	test('Searches across all projects once the user types', async ({authenticatedPage: page}) => {
		const parent = await seedProjectsWithTasks()

		await page.goto(`/tasks/${parent.id}`)
		const input = await openRelatedTasksForm(page)
		await expect(relationTaskSearchResults(page)).toHaveCount(10, {timeout: 10000})

		await input.pressSequentially('Recent task elsewhere')

		const results = relationTaskSearchResults(page)
		await expect(results).toHaveCount(1, {timeout: 10000})
		await expect(results.first()).toContainText('Other Project')
		await expect(results.first()).toContainText('Recent task elsewhere')
	})

	test('Relates a suggested task without typing', async ({authenticatedPage: page}) => {
		const parent = await seedProjectsWithTasks()

		await page.goto(`/tasks/${parent.id}`)
		await openRelatedTasksForm(page)
		await relationTaskSearchResults(page).filter({hasText: 'Suggestion 12'}).click()
		await page.locator('.task-relations select').selectOption('related')
		await page.locator('.task-relations').getByRole('button', {name: 'Add a New Task Relation'}).click()

		await expect(page.locator('.task-relations .related-tasks .task a').filter({hasText: 'Suggestion 12'}))
			.toBeVisible({timeout: 10000})
		await page.reload()
		await expect(page.locator('.task-relations .related-tasks .task a').filter({hasText: 'Suggestion 12'}))
			.toBeVisible({timeout: 10000})
	})
})
