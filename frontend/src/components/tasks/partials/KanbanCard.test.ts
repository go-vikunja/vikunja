import {describe, it, expect, vi, beforeEach} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import {normalizeTask, type TaskResponse} from '@/client/queries/tasks'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
	projectsList: vi.fn(async () => ({data: {items: [], total_pages: 1}})),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/playPop', () => ({playPopSound: vi.fn()}))

const attachments = vi.hoisted(() => ({
	fetchAttachmentUrl: vi.fn(),
	releaseAttachmentUrl: vi.fn(),
}))

vi.mock('@/helpers/attachments', () => attachments)

import KanbanCard from './KanbanCard.vue'

const recurringTask = normalizeTask({id: 7, project_id: 1, title: 'Water the plants', done: false, repeat_after: 86400})

const errorHandler = vi.fn()

function mountCard(task: TaskResponse) {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}}],
	})
	return shallowMount(KanbanCard, {
		props: {task, projectId: 1},
		global: {
			plugins: [
				createPinia(),
				router,
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			mocks: {$t: (key: string) => key},
			config: {errorHandler},
		},
	})
}

describe('KanbanCard', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		sdk.patchTasksRead.mockReset()
		attachments.fetchAttachmentUrl.mockReset()
		errorHandler.mockReset()
	})

	// The ctrl click itself is covered by e2e project-view-kanban.spec.ts, which reads the task back from the API.
	it('announces a completed recurring task to the board', async () => {
		sdk.patchTasksRead.mockResolvedValue({data: {...recurringTask, done: true}})
		const wrapper = mountCard(recurringTask)

		await wrapper.get('.task').trigger('click', {ctrlKey: true})
		await flushPromises()

		expect(wrapper.emitted('taskCompletedRecurring')).toHaveLength(1)
	})

	it('shows no cover when the cover image fails to load', async () => {
		attachments.fetchAttachmentUrl.mockRejectedValue({status: 502, detail: 'Bad Gateway'})
		const wrapper = mountCard(normalizeTask({
			id: 8,
			project_id: 1,
			title: 'Has a cover',
			cover_image_attachment_id: 3,
			attachments: [{id: 3, task_id: 8, file: {name: 'cover.jpg'}}],
		}))
		await flushPromises()

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledOnce()
		expect(errorHandler).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
	})
})
