import {type Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from './prepareProjects'

async function selectSortInList(page: Page, optionLabel: string) {
	await page.locator('.filter-container').getByRole('button', {name: 'Sort', exact: true}).click()
	await page.getByLabel('Sort by').selectOption({label: optionLabel})
	await page.getByRole('button', {name: 'Apply sort'}).click()
}

async function navigateViaSidebar(page: Page, projectTitle: string) {
	await page.locator('.menu-list .list-menu-link', {
		has: page.locator('.project-menu-title', {hasText: new RegExp(`^${projectTitle}$`)}),
	}).first().click()
}

test.describe('Sort persistence across sidebar navigation (#2753)', () => {
	test('List view: sort persists after navigating to another project and back', async ({authenticatedPage: page}) => {
		const projects = await createProjects(2)
		const [projectA, projectB] = projects
		await TaskFactory.create(3, {
			id: '{increment}',
			project_id: projectA.id,
			title: 'Task {increment}',
		})

		const listViewA = projectA.views[0].id
		await page.goto(`/projects/${projectA.id}/${listViewA}`)
		await expect(page).not.toHaveURL(/sort=/)

		await selectSortInList(page, 'Due date (Earliest first)')
		await expect(page).toHaveURL(/sort=due_date:asc/)

		await navigateViaSidebar(page, projectB.title)
		await expect(page).toHaveURL(new RegExp(`/projects/${projectB.id}/`))

		await navigateViaSidebar(page, projectA.title)
		await expect(page).toHaveURL(new RegExp(`/projects/${projectA.id}/`))
		await expect(page).toHaveURL(/sort=due_date:asc/)
	})

	test('List view: explicit URL sort wins over stored sort', async ({authenticatedPage: page}) => {
		const projects = await createProjects(1)
		const listView = projects[0].views[0].id

		// Seed the store with one sort by visiting with it set.
		await page.goto(`/projects/${projects[0].id}/${listView}?sort=due_date:asc`)
		await expect(page).toHaveURL(/sort=due_date:asc/)

		// Visit a URL that explicitly sets a different sort — that should win.
		await page.goto(`/projects/${projects[0].id}/${listView}?sort=priority:desc`)
		await expect(page).toHaveURL(/sort=priority:desc/)
	})
})
