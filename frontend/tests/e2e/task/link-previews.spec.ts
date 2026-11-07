import {writeFile} from 'node:fs/promises'
import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskCommentFactory} from '../../factories/task_comment'

const url = 'https://example.com/article'
const description = `<p><a href="${url}">${url}</a></p>`

test.beforeEach(async ({authenticatedPage: page}) => {
	await ProjectFactory.create(1)
	await TaskFactory.create(1, {id: 1, description})
	await TaskCommentFactory.create(1, {id: 1, task_id: 1, comment: description})
	await page.route('https://example.com/preview.svg', route => route.fulfill({
		contentType: 'image/svg+xml',
		body: '<svg xmlns="http://www.w3.org/2000/svg" width="160" height="100"><rect width="160" height="100" fill="#dbeafe"/><circle cx="80" cy="50" r="24" fill="#3b82f6"/></svg>',
	}))
	await page.route('**/api/v2/link-preview?**', route => route.fulfill({
		json: {
			url,
			title: 'An example article',
			description: 'A useful description of this page.',
			site_name: 'Example',
			image: 'https://example.com/preview.svg',
			favicon: '',
		},
	}))
})

test('measure unchanged editor text layout', async ({authenticatedPage: page}) => {
	await page.goto('/tasks/1')
	const editor = page.locator('.description .tiptap__editor .ProseMirror')
	await expect(editor).toContainText(url)
	const metrics = await editor.evaluate(element => {
		const rect = element.getBoundingClientRect()
		const style = getComputedStyle(element)
		return {
			width: rect.width,
			height: rect.height,
			display: style.display,
			gap: style.gap,
			margin: style.margin,
			padding: style.padding,
			lineHeight: style.lineHeight,
			fontSize: style.fontSize,
		}
	})
	await writeFile(process.env.LINK_PREVIEW_METRICS || '/tmp/link-preview-metrics.json', JSON.stringify(metrics))
})

test('previews standalone URLs in descriptions and comments without storing cards', async ({authenticatedPage: page}) => {
	await page.goto('/tasks/1')
	await expect(page.locator('.description .link-preview')).toContainText('An example article')
	await expect(page.locator('#comment-1 .link-preview')).toContainText('A useful description')
	await expect(page.locator('.description .link-preview__image')).toHaveJSProperty('naturalWidth', 160)
	await page.screenshot({path: '/tmp/link-preview-desktop.png', fullPage: true})
	await page.locator('.description .done-edit').click()
	await expect(page.locator('.description .link-preview')).toHaveCount(0)
	const editor = page.locator('.description .ProseMirror')
	await editor.press('End')
	await editor.press('Enter')
	await editor.pressSequentially('Additional text')
	await page.getByTestId('saveEditor').click()
	await expect(page.locator('.description .link-preview')).toBeVisible()
	await page.reload()
	await expect(page.locator('.description .link-preview')).toContainText('An example article')
	await expect(page.locator('.description .ProseMirror')).toContainText('Additional text')
	await expect(page.locator('.description .ProseMirror .link-preview')).toHaveCount(0)
})

test('skips inline, named, internal and unsafe links', async ({authenticatedPage: page}) => {
	const internalUrl = new URL('/projects/1', process.env.BASE_URL || 'http://127.0.0.1:4173').href
	await TaskCommentFactory.create(1, {id: 1, task_id: 1, comment: '<p>No links</p>'})
	await TaskFactory.create(1, {
		id: 1,
		description: `<p>Read <a href="${url}">${url}</a> today.</p>` +
			`<p><a href="${url}">Named link</a></p>` +
			`<p><a href="${internalUrl}">${internalUrl}</a></p>` +
			'<p><a href="ftp://example.com/file">ftp://example.com/file</a></p>',
	})
	const requests: string[] = []
	page.on('request', request => {
		if (request.url().includes('/link-preview?')) requests.push(request.url())
	})
	await page.goto('/tasks/1')
	await expect(page.locator('.description .ProseMirror')).toContainText('Named link')
	await expect(page.locator('.description .link-preview')).toHaveCount(0)
	expect(requests).toEqual([])
})

test('pastes a standalone URL and persists only the original content', async ({authenticatedPage: page, apiContext, userToken}) => {
	await TaskFactory.create(1, {id: 1, description: '<p>Introduction</p>'})
	await page.goto('/tasks/1')
	await page.locator('.description .done-edit').click()
	const editor = page.locator('.description .ProseMirror')
	await editor.click()
	await editor.press('ControlOrMeta+End')
	await editor.press('Enter')
	await page.evaluate(text => navigator.clipboard.writeText(text), url)
	await editor.press('ControlOrMeta+v')
	await expect(editor.locator('a')).toHaveAttribute('href', url)
	await page.getByTestId('saveEditor').click()
	await expect(page.locator('.description .link-preview')).toContainText('An example article')
	await page.reload()
	await expect(page.locator('.description .link-preview')).toContainText('An example article')
	const response = await apiContext.get('tasks/1', {
		headers: {Authorization: `Bearer ${userToken}`},
	})
	expect(response.ok()).toBeTruthy()
	const task = await response.json()
	expect(task.description).toContain(url)
	expect(task.description).not.toContain('An example article')
	expect(task.description).not.toContain('link-preview')
})

test('unavailable previews leave the original link usable', async ({authenticatedPage: page}) => {
	await page.route('**/api/v2/link-preview?**', route => route.fulfill({status: 502, json: {detail: 'Unavailable'}}))
	await page.goto('/tasks/1')
	await expect(page.locator('.description .ProseMirror a')).toHaveAttribute('href', url)
	await expect(page.locator('.description .link-preview')).toHaveCount(0)
	await expect(page.locator('.notices .notification.is-danger')).toHaveCount(0)
})

test('cards fit on a narrow screen', async ({authenticatedPage: page}) => {
	await page.setViewportSize({width: 390, height: 844})
	await page.goto('/tasks/1')
	const card = page.locator('.description .link-preview')
	await expect(card).toBeVisible()
	const box = await card.boundingBox()
	expect(box).not.toBeNull()
	expect(box!.x).toBeGreaterThanOrEqual(0)
	expect(box!.x + box!.width).toBeLessThanOrEqual(390)
	await page.screenshot({path: '/tmp/link-preview-mobile.png', fullPage: true})
})
