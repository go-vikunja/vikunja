import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {createProjects} from '../project/prepareProjects'
import {TaskFactory} from '../../factories/task'
import {LabelTaskFactory} from '../../factories/label_task'
import {LabelFactory} from '../../factories/labels'
import {SavedFilterFactory} from '../../factories/saved_filter'

async function openFilters(page: Page) {
	await page.locator('.filter-container button').filter({hasText: 'Filters'}).click()
}

test.beforeEach(async ({authenticatedPage: page, currentUser}) => {
	await createProjects()
	await TaskFactory.create(2, {
		project_id: 1,
		title: (i: number) => `Task ${i}`,
	})
	await LabelFactory.create(1, {id: 1, title: 'Important', created_by_id: currentUser.id})
	await LabelTaskFactory.create(1, {
		task_id: 1,
		label_id: 1,
	})
	await page.goto('/projects/1/1')
	await openFilters(page)
	await expect(page.locator('.filter-popup')).toBeVisible()
})

test('selects a label in Simple and preserves the query across modes and reload', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	await expect(popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})).toHaveAttribute('aria-selected', 'true')
	await popup.getByRole('combobox', {
		name: 'Labels',
		exact: true,
	}).pressSequentially('Important')
	await popup.getByRole('option', {name: /Important/}).click()
	await expect(page.locator('.tasks')).toContainText('Task 1')
	await expect(page.locator('.tasks')).not.toContainText('Task 2')
	await popup.getByRole('tab', {
		name: 'Advanced',
		exact: true,
	}).click()
	await expect(popup.locator('.ProseMirror')).toContainText('labels in Important')
	await page.reload()
	await expect(page.locator('.tasks')).not.toContainText('Task 2')
	await openFilters(page)
	await expect(page.locator('.filter-popup .ProseMirror')).toContainText('labels in Important')
})

test('keeps an unsupported query in Advanced until explicitly cleared', async ({authenticatedPage: page}) => {
	await page.goto('/projects/1/1?filter=priority%20%3E%203')
	await openFilters(page)
	const popup = page.locator('.filter-popup')
	await expect(popup.getByRole('tab', {
		name: 'Advanced',
		exact: true,
	})).toHaveAttribute('aria-selected', 'true')
	await expect(popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})).toHaveAttribute('aria-disabled', 'true')
	await expect(popup.getByText('This filter is too advanced for Simple mode.')).toBeVisible()
	await popup.getByRole('button', {name: 'Clear and use Simple'}).click()
	const simpleTab = popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})
	await expect(simpleTab).toHaveAttribute('aria-selected', 'true')
	await expect(simpleTab).toBeFocused()
	await expect(page).not.toHaveURL(/filter=/)
})

test('does not apply an Advanced popup query when the popup is closed without Show results', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	await popup.getByRole('tab', {
		name: 'Advanced',
		exact: true,
	}).click()
	await popup.locator('.ProseMirror').fill('done = true')
	await page.getByRole('button', {
		name: 'Close dialog',
		exact: true,
	}).click()
	await expect(popup).toBeHidden()
	expect(new URL(page.url()).searchParams.get('filter')).toBeNull()
	await expect(page.locator('.tasks')).toContainText('Task 1')
	await expect(page.locator('.tasks')).toContainText('Task 2')
})

test('clears an Advanced plain-text search from the popup', async ({authenticatedPage: page}) => {
	await page.goto('/projects/1/1?s=Task%201')
	await expect(page.locator('.tasks')).not.toContainText('Task 2')
	await openFilters(page)
	const popup = page.locator('.filter-popup')
	await popup.getByRole('tab', {
		name: 'Advanced',
		exact: true,
	}).click()
	const editor = popup.locator('.ProseMirror')
	await expect(editor).toHaveText('Task 1')
	await editor.press('ControlOrMeta+a')
	await editor.press('Backspace')
	await popup.getByRole('button', {
		name: 'Show results',
		exact: true,
	}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('s')).toBeNull()
	await expect(page.locator('.tasks')).toContainText('Task 1')
	await expect(page.locator('.tasks')).toContainText('Task 2')
})

test('keeps Advanced input while typing on the new saved filter page', async ({authenticatedPage: page}) => {
	await page.goto('/filters/new')
	await page.getByRole('tab', {
		name: 'Advanced',
		exact: true,
	}).click()
	const editor = page.locator('.filter-editor .ProseMirror')
	await editor.press('ControlOrMeta+a')
	await editor.pressSequentially('done = true')
	await expect(editor).toHaveText('done = true')
})

test('keeps search separate and applies it on blur', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	const search = popup.getByRole('searchbox', {name: 'Search tasks'})
	await search.fill('Task 1')
	await expect(page.locator('.tasks')).toContainText('Task 2')
	await popup.getByRole('checkbox', {name: 'Show completed tasks'}).focus()
	await expect(page.locator('.tasks')).toContainText('Task 1')
	await expect(page.locator('.tasks')).not.toContainText('Task 2')
	await page.reload()
	await expect(page.locator('.tasks')).toContainText('Task 1')
	await expect(page.locator('.tasks')).not.toContainText('Task 2')
})

test('creates saved filters in Simple and reloads them for editing', async ({authenticatedPage: page, apiContext, userToken}) => {
	await page.goto('/filters/new')
	await page.locator('input#Title').fill('Simple saved filter')
	await page.getByRole('combobox', {
		name: 'Labels',
		exact: true,
	}).pressSequentially('Important')
	await page.getByRole('option', {name: /Important/}).click()
	await page.getByRole('combobox', {
		name: 'Projects',
		exact: true,
	}).pressSequentially('First')
	await page.getByRole('option', {name: /First Project/}).click()
	const created = page.waitForResponse(response => response.url().endsWith('/api/v2/filters') && response.request().method() === 'POST')
	await page.locator('button.is-primary.is-fullwidth').click()
	const response = await created
	expect(response.ok()).toBe(true)
	const {id} = await response.json()
	await page.goto(`/projects/${-id - 1}/settings/edit`)
	await expect(page.getByRole('combobox', {
		name: 'Projects',
		exact: true,
	})).toBeVisible()
	await expect(page.locator('.filter-editor')).toContainText('Important')
	const stored = await apiContext.get(`filters/${id}`, {headers: {Authorization: `Bearer ${userToken}`}})
	expect(stored.ok()).toBe(true)
	expect((await stored.json()).filters.filter).toBe('done = false && labels in 1 && project in 1')
})

test('preserves unknown label and project ids when another field changes', async ({authenticatedPage: page, apiContext, userToken}) => {
	await SavedFilterFactory.create(1, {filters: JSON.stringify({
		filter: 'labels in 999 && project in 999',
		s: '',
		filter_include_nulls: false,
	})})
	await page.goto('/projects/-2/settings/edit')
	const editor = page.locator('.filter-editor')
	await expect(editor).toContainText('Unknown label #999')
	await expect(editor).toContainText('Unknown project #999')
	await editor.getByText('Show completed tasks', {exact: true}).click()
	const saved = page.waitForResponse(response => response.url().endsWith('/api/v2/filters/1') && response.request().method() === 'PUT')
	await page.locator('footer.card-footer .button').filter({hasText: /^Save$/}).click()
	expect((await saved).ok()).toBe(true)
	const stored = await apiContext.get('filters/1', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await stored.json()).filters.filter).toBe('done = false && labels in 999 && project in 999')
	await page.goto('/projects/-2/settings/edit')
	await expect(editor).toContainText('Unknown label #999')
	await expect(editor).toContainText('Unknown project #999')
	await editor.locator('.multiselect').filter({hasText: 'Unknown label #999'}).getByRole('button', {name: 'Remove this item'}).click()
	await expect(editor).not.toContainText('Unknown label #999')
})

test('opens the default open tasks filter in Simple without rewriting it', async ({authenticatedPage: page, apiContext, userToken, currentUser}) => {
	const filter = `done = false && assignees = ${currentUser.username}`
	await SavedFilterFactory.create(1, {filters: JSON.stringify({
		filter,
		s: '',
		filter_include_nulls: false,
	})})
	await page.goto('/projects/-2/settings/edit')
	const editor = page.locator('.filter-editor')
	await expect(editor.getByRole('group', {name: 'Assignees'})).toContainText(currentUser.username)
	await expect(editor.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})).toHaveAttribute('aria-selected', 'true')
	await expect(editor.getByText('This filter is too advanced for Simple mode.')).toBeHidden()
	await page.reload()
	await expect(editor.getByRole('group', {name: 'Assignees'})).toContainText(currentUser.username)
	const stored = await apiContext.get('filters/1', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await stored.json()).filters.filter).toBe(filter)
})

test('saves the typed search when Enter is pressed in a saved filter', async ({authenticatedPage: page, apiContext, userToken}) => {
	await SavedFilterFactory.create(1)
	await page.goto('/projects/-2/settings/edit')
	const editor = page.locator('.filter-editor')
	await expect(editor.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})).toHaveAttribute('aria-selected', 'true')
	const saved = page.waitForResponse(response => response.url().endsWith('/api/v2/filters/1') && response.request().method() === 'PUT')
	const search = editor.getByRole('searchbox', {name: 'Search tasks'})
	await search.fill('Task 1')
	await search.press('Enter')
	expect((await saved).ok()).toBe(true)
	const stored = await apiContext.get('filters/1', {headers: {Authorization: `Bearer ${userToken}`}})
	expect((await stored.json()).filters.s).toBe('Task 1')
})

test('applies a due preset, includes overdue tasks, and clears the range', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	await popup.getByRole('button', {
		name: 'Select a date range',
		exact: true,
	}).click()
	const picker = page.locator('.datepicker-with-range.is-open')
	await picker.getByRole('button', {
		name: 'Today',
		exact: true,
	}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date > now/d && due_date < now/d+1d')
	await popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	}).click()
	await popup.getByText('Show overdue tasks', {exact: true}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now/d+1d')
	await popup.getByRole('button', {
		name: 'Clear',
		exact: true,
	}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now')
	await popup.getByRole('button', {
		name: 'Select a date range',
		exact: true,
	}).click()
	await picker.getByRole('button', {
		name: 'Today',
		exact: true,
	}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now/d+1d')
	await picker.getByRole('button', {
		name: 'Custom',
		exact: true,
	}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now')
	await picker.getByLabel('From', {exact: true}).fill('now/d')
	await picker.getByLabel('To', {exact: true}).pressSequentially('now/d+2d')
	await expect(popup.getByRole('tab', {name: 'Simple', exact: true})).toHaveAttribute('aria-selected', 'true')
	await expect(picker.getByLabel('To', {exact: true})).toHaveValue('now/d+2d')
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now/d+2d')

})

test('labels an overdue range by its end and restores a range when overdue is unticked after reload', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	await popup.getByRole('button', {
		name: 'Select a date range',
		exact: true,
	}).click()
	await page.locator('.datepicker-with-range.is-open').getByRole('button', {
		name: 'Today',
		exact: true,
	}).click()
	await popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	}).click()
	await popup.getByText('Show overdue tasks', {exact: true}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now/d+1d')
	const until = popup.getByRole('button', {
		name: 'until now/d+1d',
		exact: true,
	})
	await expect(until).toBeVisible()
	await page.reload()
	await openFilters(page)
	await expect(until).toBeVisible()
	await popup.getByText('Show overdue tasks', {exact: true}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date > now && due_date < now/d+1d')
})

test('drops the due filter when overdue is unticked after typing now as the end', async ({authenticatedPage: page}) => {
	const popup = page.locator('.filter-popup')
	await popup.getByRole('button', {
		name: 'Select a date range',
		exact: true,
	}).click()
	const picker = page.locator('.datepicker-with-range.is-open')
	await picker.getByRole('button', {
		name: 'Custom',
		exact: true,
	}).click()
	await picker.getByLabel('To', {exact: true}).pressSequentially('now')
	await popup.getByRole('tab', {
		name: 'Simple',
		exact: true,
	}).click()
	await popup.getByText('Show overdue tasks', {exact: true}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter')).toBe('due_date < now')
	await popup.getByText('Show overdue tasks', {exact: true}).click()
	await expect.poll(() => new URL(page.url()).searchParams.get('filter') ?? '').not.toContain('due_date')
})

test('saves a view filter through the shared Simple editor', async ({authenticatedPage: page, apiContext, userToken}) => {
	await page.goto('/projects/1/settings/views')
	await page.getByRole('button', {
		name: 'Edit this view',
		exact: true,
	}).first().click()
	const form = page.locator('table form')
	await expect(form.getByRole('tab', {
		name: 'Simple',
		exact: true,
	})).toHaveAttribute('aria-selected', 'true')
	await expect(form.getByRole('combobox', {
		name: 'Projects',
		exact: true,
	})).toHaveCount(0)
	await form.getByRole('combobox', {
		name: 'Labels',
		exact: true,
	}).pressSequentially('Important')
	await form.getByRole('option', {name: /Important/}).click()
	const saved = page.waitForResponse(response => response.url().endsWith('/api/v2/projects/1/views/1') && response.request().method() === 'PUT')
	await form.getByRole('button', {
		name: 'Save',
		exact: true,
	}).click()
	expect((await saved).ok()).toBe(true)
	const stored = await apiContext.get('projects/1/views', {headers: {Authorization: `Bearer ${userToken}`}})
	const savedView = (await stored.json()).find((view: {
		id: number
	}) => view.id === 1)
	expect(savedView.filter.filter).toBe('labels in 1')
	await page.reload()
	await page.getByRole('row').filter({hasText: savedView.title})
		.getByRole('button', {
			name: 'Edit this view',
			exact: true,
		}).click()
	await expect(page.locator('table form .filter-editor')).toContainText('Important')
})
