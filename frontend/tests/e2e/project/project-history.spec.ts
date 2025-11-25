import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'

test.describe('Project History', () => {
	// FIXME: API timeout waiting for /projects response - likely related to project_views table seeding issues
	test.skip('should show a project history on the home page', async ({authenticatedPage: page}) => {
		test.setTimeout(60000)
		const projects = await ProjectFactory.create(7)
		await ProjectViewFactory.truncate()
		await Promise.all(projects.map(p => ProjectViewFactory.create(1, {
			id: p.id,
			project_id: p.id,
		}, false)))

		const loadProjectArrayPromise = page.waitForResponse(response =>
			response.url().includes('/projects') && !response.url().includes('/projects/'),
		)
		await page.goto('/')
		await loadProjectArrayPromise
		await expect(page.locator('body')).not.toContainText('Last viewed')

		for (let i = 0; i < projects.length; i++) {
			const loadProjectPromise = page.waitForResponse(response =>
				response.url().includes(`/projects/${projects[i].id}`) && response.request().method() === 'GET',
			)
			await page.goto(`/projects/${projects[i].id}/${projects[i].id}`)
			await loadProjectPromise
			// Wait for history to be saved to localStorage
			await page.waitForFunction(
				(projectId) => {
					const history = JSON.parse(localStorage.getItem('projectHistory') || '[]')
					return history.some((h: any) => h.id === projectId)
				},
				projects[i].id,
			)
		}

		// Not using goto here to work around the redirect issue fixed in #1337
		const loadProjectArrayPromise2 = page.waitForResponse(response =>
			response.url().includes('/projects') && !response.url().includes('/projects/'),
		)
		await page.locator('nav.menu.top-menu a').filter({hasText: 'Overview'}).click()
		await loadProjectArrayPromise2

		await expect(page.locator('body')).toContainText('Last viewed')
		await expect(page.locator('.project-grid')).not.toContainText(projects[0].title)
		await expect(page.locator('.project-grid')).toContainText(projects[1].title)
		await expect(page.locator('.project-grid')).toContainText(projects[2].title)
		await expect(page.locator('.project-grid')).toContainText(projects[3].title)
		await expect(page.locator('.project-grid')).toContainText(projects[4].title)
		await expect(page.locator('.project-grid')).toContainText(projects[5].title)
		await expect(page.locator('.project-grid')).toContainText(projects[6].title)
	})
})
