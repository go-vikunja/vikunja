import {describe, it, expect, vi, beforeEach} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {Task as ITask} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/playPop', () => ({playPopSound: vi.fn()}))
vi.mock('@/composables/useProjects', () => ({useProjects: () => ({projects: {}})}))
vi.mock('@/composables/useCurrentProject', () => ({useCurrentProject: () => ({currentProject: {value: null}})}))
vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

import SingleTaskInProject from './SingleTaskInProject.vue'

const theTask = {id: 7, project_id: 1, title: 'Water the plants', done: true, is_favorite: false} as ITask

function deferred() {
	let resolve!: (value: unknown) => void
	const promise = new Promise(done => resolve = done)
	return {promise, resolve}
}

function mountRow() {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}},
			{path: '/projects/:projectId', name: 'project.index', component: {render: () => null}},
		],
	})
	return shallowMount(SingleTaskInProject, {
		props: {theTask},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			directives: {tooltip: {}},
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('SingleTaskInProject', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		sdk.patchTasksRead.mockReset()
	})

	it('shows the row as loading while marking it undone is in flight', async () => {
		const update = deferred()
		sdk.patchTasksRead.mockReturnValue(update.promise)
		const wrapper = mountRow()

		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', false)
		await flushPromises()

		expect(wrapper.get('.task').classes()).toContain('is-loading')

		update.resolve({data: {...theTask, done: false}})
		await flushPromises()

		expect(wrapper.get('.task').classes()).not.toContain('is-loading')
	})

	it('shows the row as loading while favouring it is in flight', async () => {
		const favorite = deferred()
		sdk.patchTasksRead.mockReturnValue(favorite.promise)
		const wrapper = mountRow()

		await wrapper.get('.favorite').trigger('click')
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledWith(expect.objectContaining({
			path: {task: 7},
			body: [expect.objectContaining({path: '/is_favorite', value: true})],
		}))
		expect(wrapper.get('.task').classes()).toContain('is-loading')

		favorite.resolve({data: {...theTask, is_favorite: true}})
		await flushPromises()

		expect(wrapper.get('.task').classes()).not.toContain('is-loading')
	})
})
