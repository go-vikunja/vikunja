import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
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

class FakeIntersectionObserver {
	static instances: FakeIntersectionObserver[] = []
	observing = false
	target: Element | undefined

	constructor(
		private callback: IntersectionObserverCallback,
		public options?: IntersectionObserverInit,
	) {
		FakeIntersectionObserver.instances.push(this)
	}

	observe(target: Element) {
		this.observing = true
		this.target = target
	}

	disconnect() {
		this.observing = false
	}

	report(isIntersecting: boolean) {
		if (!this.observing) {
			return
		}
		this.callback([{isIntersecting} as IntersectionObserverEntry], this as unknown as IntersectionObserver)
	}
}

async function scrollCard(isIntersecting: boolean) {
	FakeIntersectionObserver.instances.forEach(observer => observer.report(isIntersecting))
	await flushPromises()
}

const recurringTask = normalizeTask({id: 7, project_id: 1, title: 'Water the plants', done: false, repeat_after: 86400})

const taskWithCover = normalizeTask({
	id: 8,
	project_id: 1,
	title: 'Has a cover',
	cover_image_attachment_id: 3,
	attachments: [
		{id: 3, task_id: 8, file: {name: 'cover.jpg', mime: 'image/jpeg'}},
		{id: 4, task_id: 8, file: {name: 'other.png', mime: 'image/png'}},
	],
})

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
		FakeIntersectionObserver.instances = []
		vi.stubGlobal('IntersectionObserver', FakeIntersectionObserver)
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	// The ctrl click itself is covered by e2e project-view-kanban.spec.ts, which reads the task back from the API.
	it('announces a completed recurring task to the board', async () => {
		sdk.patchTasksRead.mockResolvedValue({data: {...recurringTask, done: true}})
		const wrapper = mountCard(recurringTask)

		await wrapper.get('.task').trigger('click', {ctrlKey: true})
		await flushPromises()

		expect(wrapper.emitted('taskCompletedRecurring')).toHaveLength(1)
	})

	it('does not load the cover while the card is off screen', async () => {
		const wrapper = mountCard(taskWithCover)
		await scrollCard(false)

		expect(attachments.fetchAttachmentUrl).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
	})

	it('loads the cover once when the card scrolls into view', async () => {
		attachments.fetchAttachmentUrl.mockResolvedValue('blob:cover-3')
		const wrapper = mountCard(taskWithCover)

		await scrollCard(true)

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledExactlyOnceWith({id: 3, task_id: 8}, 'lg')
		expect(wrapper.get('img').attributes('src')).toBe('blob:cover-3')

		await scrollCard(false)
		await scrollCard(true)

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledOnce()
		expect(wrapper.get('img').attributes('src')).toBe('blob:cover-3')
	})

	it('follows cover changes on a card that has been visible', async () => {
		attachments.fetchAttachmentUrl.mockImplementation(async ({id}: {id: number}) => `blob:cover-${id}`)
		const wrapper = mountCard(taskWithCover)
		await scrollCard(true)
		await scrollCard(false)

		await wrapper.setProps({task: {...taskWithCover, cover_image_attachment_id: 4}})
		await flushPromises()

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledTimes(2)
		expect(wrapper.get('img').attributes('src')).toBe('blob:cover-4')

		await wrapper.setProps({task: {...taskWithCover, cover_image_attachment_id: 0}})
		await flushPromises()

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledTimes(2)
		expect(wrapper.find('img').exists()).toBe(false)
	})

	it('waits for the card to scroll into view before loading a changed cover', async () => {
		attachments.fetchAttachmentUrl.mockImplementation(async ({id}: {id: number}) => `blob:cover-${id}`)
		const wrapper = mountCard(taskWithCover)
		await scrollCard(false)

		await wrapper.setProps({task: {...taskWithCover, cover_image_attachment_id: 4}})
		await flushPromises()

		expect(attachments.fetchAttachmentUrl).not.toHaveBeenCalled()

		await scrollCard(true)

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledExactlyOnceWith({id: 4, task_id: 8}, 'lg')
	})

	it('shows no cover when the cover image fails to load', async () => {
		attachments.fetchAttachmentUrl.mockRejectedValue({status: 502, detail: 'Bad Gateway'})
		const wrapper = mountCard(taskWithCover)
		await scrollCard(true)

		expect(attachments.fetchAttachmentUrl).toHaveBeenCalledOnce()
		expect(errorHandler).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
	})

	it('does not request a cover preview for a file that is not an image', async () => {
		const wrapper = mountCard({
			...taskWithCover,
			attachments: [{id: 3, task_id: 8, file: {name: 'cover.jpg', mime: 'text/plain'}}],
		})
		await scrollCard(true)

		expect(attachments.fetchAttachmentUrl).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
	})

	it('stops watching for visibility when it is unmounted', async () => {
		const wrapper = mountCard(taskWithCover)
		const [observer] = FakeIntersectionObserver.instances
		expect(observer.observing).toBe(true)
		expect(observer.target).toBe(wrapper.element)
		expect(observer.options).toEqual({
			rootMargin: '200px',
			scrollMargin: '200px',
		})

		wrapper.unmount()
		await scrollCard(true)

		expect(observer.observing).toBe(false)
		expect(attachments.fetchAttachmentUrl).not.toHaveBeenCalled()
	})
})
