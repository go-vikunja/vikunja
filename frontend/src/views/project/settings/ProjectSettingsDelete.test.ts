import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {ProjectResponse} from '@/client/queries/projects'

const sdk = vi.hoisted(() => ({
	tasksList: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)

const {projects} = vi.hoisted(() => ({projects: {} as Record<number, ProjectResponse>}))

vi.mock('@/composables/useProjects', () => ({
	useProjects: () => ({
		projects,
		getChildProjects: () => [],
	}),
}))

import ProjectSettingsDelete from '@/views/project/settings/ProjectSettingsDelete.vue'
import Modal from '@/components/misc/Modal.vue'
import Loading from '@/components/misc/Loading.vue'
import testid from '@/directives/testid'
import enMessages from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en: enMessages}})

async function mountModal() {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{
			path: '/projects/:projectId/settings/delete',
			name: 'project.settings.delete',
			component: ProjectSettingsDelete,
		}],
	})
	await router.push('/projects/42/settings/delete')
	await router.isReady()

	const wrapper = mount(ProjectSettingsDelete, {
		global: {
			plugins: [i18n, router, [VueQueryPlugin, {queryClient: new QueryClient()}]],
			components: {Modal},
			directives: {cy: testid},
			stubs: {
				BaseButton: {template: '<a v-bind="$attrs"><slot /></a>'},
				XButton: {template: '<button type="button" v-bind="$attrs"><slot /></button>'},
				Icon: true,
			},
		},
		attachTo: document.body,
	})
	await flushPromises()
	return wrapper
}

describe('ProjectSettingsDelete', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		setActivePinia(createPinia())
		document.body.innerHTML = ''
		projects[42] = {id: 42, title: 'Test'} as ProjectResponse
		sdk.tasksList.mockReset()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		document.body.innerHTML = ''
	})

	it('shows the loader while the task count is pending', async () => {
		sdk.tasksList.mockReturnValue(new Promise(() => {}))

		wrapper = await mountModal()

		expect(wrapper.findComponent(Loading).exists()).toBe(true)
		expect(document.querySelector('dialog.modal-dialog')?.textContent)
			.not.toContain('does not contain any tasks')
	})

	it('shows the task count once the query resolved', async () => {
		sdk.tasksList.mockResolvedValue({data: {items: [], total: 3, page: 1, per_page: 1, total_pages: 3}})

		wrapper = await mountModal()

		expect(wrapper.findComponent(Loading).exists()).toBe(false)
		expect(document.querySelector('dialog.modal-dialog')?.textContent)
			.toContain('irrevocably remove approx. 3 tasks')
	})
})
