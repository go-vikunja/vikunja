import {describe, it, expect, vi, beforeEach} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {Task} from '@/client/generated'

const {patchTasksRead} = vi.hoisted(() => ({patchTasksRead: vi.fn()}))

vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<object>(),
	patchTasksRead,
	projectsList: vi.fn(async () => ({data: {items: [], total_pages: 1}})),
}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/playPop', () => ({playPopSound: vi.fn()}))

import KanbanCard from './KanbanCard.vue'

const recurringTask: Task = {id: 7, project_id: 1, title: 'Water the plants', done: false, repeat_after: 86400}

function mountCard(task: Task) {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}}],
	})
	return shallowMount(KanbanCard, {
		props: {task, projectId: 1},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('KanbanCard', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		patchTasksRead.mockReset()
	})

	it('marks the task done through the api on ctrl click', async () => {
		patchTasksRead.mockResolvedValue({data: {...recurringTask, done: true}})
		const wrapper = mountCard(recurringTask)

		await wrapper.get('.task').trigger('click', {ctrlKey: true})
		await flushPromises()

		expect(patchTasksRead).toHaveBeenCalledWith(expect.objectContaining({
			path: {task: 7},
			body: expect.arrayContaining([expect.objectContaining({path: '/done', value: true})]),
		}))
		expect(wrapper.emitted('taskCompletedRecurring')).toHaveLength(1)
	})
})
