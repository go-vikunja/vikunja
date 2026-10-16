import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {ProjectResponse} from '@/client/queries/projects'

const sdk = vi.hoisted(() => ({
	projectsCount: vi.fn(),
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
		sdk.projectsCount.mockReset()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		document.body.innerHTML = ''
	})

	it('shows the loader while the counts are pending', async () => {
		sdk.projectsCount.mockReturnValue(new Promise(() => {}))

		wrapper = await mountModal()

		expect(wrapper.findComponent(Loading).exists()).toBe(true)
		expect(document.querySelector('dialog.modal-dialog')?.textContent).not.toContain('Moves this project')
	})

	it('tells what moves to the trash once the counts resolved', async () => {
		sdk.projectsCount.mockResolvedValue({data: {sub_projects: 0, tasks: 3, time_entries: 0, time_entry_authors: 0}})

		wrapper = await mountModal()

		expect(sdk.projectsCount).toHaveBeenCalledWith(expect.objectContaining({path: {id: 42}}))
		expect(wrapper.findComponent(Loading).exists()).toBe(false)
		const text = document.querySelector('dialog.modal-dialog')?.textContent
		expect(text).toContain('Moves this project and 3 tasks to the trash.')
		expect(text).not.toContain('time entries')
	})

	it('mentions sub-projects and time entries', async () => {
		sdk.projectsCount.mockResolvedValue({data: {sub_projects: 2, tasks: 1, time_entries: 4, time_entry_authors: 2}})

		wrapper = await mountModal()

		const text = document.querySelector('dialog.modal-dialog')?.textContent
		expect(text).toContain('Moves this project, 2 sub-projects and 1 task to the trash.')
		expect(text).toContain('4 time entries from 2 people will be kept but detached')
	})
})
