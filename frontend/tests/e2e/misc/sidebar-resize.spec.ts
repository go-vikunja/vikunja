import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews} from '../project/prepareProjects'
import type {APIRequestContext} from '@playwright/test'

async function seedTasks(apiContext: APIRequestContext, numberOfTasks = 5, startDueDate = new Date()) {
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

test.describe('Sidebar Resize', () => {
	test('should NOT reload tasks when resizing the sidebar', async ({authenticatedPage: page, apiContext}) => {
		// Set a desktop viewport to ensure resize handle is visible (must be > 768px)
		await page.setViewportSize({width: 1280, height: 720})

		// Create tasks
		await seedTasks(apiContext, 5)

		// Go to homepage
		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Wait for tasks to be visible
		await expect(page.locator('[data-cy="showTasks"] .card .task').first()).toBeVisible()

		// Track task API calls during resize
		let taskApiCalls = 0
		const taskCallTimestamps: string[] = []
		page.on('request', request => {
			if (request.url().includes('/tasks') && request.method() === 'GET') {
				taskApiCalls++
				taskCallTimestamps.push(`[${new Date().toISOString()}] ${request.url()}`)
				console.log(`Task API call #${taskApiCalls}: ${request.url()}`)
			}
		})

		// Wait for any pending requests to complete, then reset counter
		await page.waitForTimeout(500)
		taskApiCalls = 0

		// Find the resize handle
		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).toBeAttached()

		// Get sidebar bounding box to calculate drag position
		const sidebar = page.locator('.menu-container')
		const sidebarBox = await sidebar.boundingBox()
		expect(sidebarBox).not.toBeNull()

		// Drag from right edge of sidebar (where resize handle is)
		const startX = sidebarBox!.x + sidebarBox!.width - 2
		const startY = sidebarBox!.y + sidebarBox!.height / 2

		// Perform resize drag
		await page.mouse.move(startX, startY)
		await page.mouse.down()
		await page.mouse.move(startX + 100, startY, {steps: 10})
		await page.mouse.up()

		// Wait for settings save to complete
		await page.waitForTimeout(1500)

		// Verify no task API calls were made during resize
		if (taskApiCalls > 0) {
			console.log('Task API calls made during resize:')
			taskCallTimestamps.forEach(t => console.log(t))
		}
		expect(taskApiCalls).toBe(0)
	})

	test('should persist sidebar width after resize', async ({authenticatedPage: page, apiContext}) => {
		// Set a desktop viewport
		await page.setViewportSize({width: 1280, height: 720})

		// Create a project so homepage has content
		await seedTasks(apiContext, 1)

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Get initial sidebar width
		const sidebar = page.locator('.menu-container')
		const initialBox = await sidebar.boundingBox()
		expect(initialBox).not.toBeNull()
		const initialWidth = initialBox!.width

		// Resize sidebar
		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).toBeAttached()

		const startX = initialBox!.x + initialBox!.width - 2
		const startY = initialBox!.y + initialBox!.height / 2

		await page.mouse.move(startX, startY)
		await page.mouse.down()
		await page.mouse.move(startX + 50, startY, {steps: 10})
		await page.mouse.up()

		// Wait for settings save
		await page.waitForTimeout(1000)

		// Get new width
		const newBox = await sidebar.boundingBox()
		expect(newBox).not.toBeNull()
		expect(newBox!.width).toBeGreaterThan(initialWidth)

		// Reload page and verify width persisted
		await page.reload()
		await page.waitForLoadState('networkidle')

		const reloadedBox = await sidebar.boundingBox()
		expect(reloadedBox).not.toBeNull()
		// Allow some tolerance for rounding
		expect(Math.abs(reloadedBox!.width - newBox!.width)).toBeLessThan(5)
	})

	test('should not show resize handle on mobile', async ({authenticatedPage: page, apiContext}) => {
		// Create a project so homepage has content
		await seedTasks(apiContext, 1)

		// Set mobile viewport (below 768px breakpoint)
		await page.setViewportSize({width: 375, height: 667})

		await page.goto('/')
		await page.waitForLoadState('networkidle')

		// Resize handle should not be in DOM on mobile
		const resizeHandle = page.locator('.resize-handle')
		await expect(resizeHandle).not.toBeAttached()
	})
})
