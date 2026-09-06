import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'

import type {ProjectResponse} from '@/client/queries/projects'

const {projects} = vi.hoisted(() => ({projects: {} as Record<number, ProjectResponse>}))

vi.mock('@/composables/useProjectNavigation', () => ({
	useProjectNavigation: () => ({
		projects,
		updateProject: vi.fn(),
		loadAllProjects: vi.fn(),
	}),
}))

import ProjectSettingsArchive from '@/views/project/settings/ProjectSettingsArchive.vue'
import Modal from '@/components/misc/Modal.vue'
import testid from '@/directives/testid'
import enMessages from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en: enMessages}})

function createTestRouter() {
	return createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/projects/:projectId/settings/archive', name: 'project.settings.archive', component: ProjectSettingsArchive},
			{path: '/tasks/by/upcoming', name: 'tasks.range', component: {template: '<div />'}},
		],
	})
}

async function mountModal(path: string) {
	const errors: unknown[] = []
	const router = createTestRouter()
	await router.push(path)
	await router.isReady()

	const wrapper = mount(ProjectSettingsArchive, {
		global: {
			plugins: [i18n, router],
			components: {Modal},
			directives: {cy: testid},
			stubs: {
				BaseButton: {template: '<a v-bind="$attrs"><slot /></a>'},
				XButton: {template: '<button type="button" v-bind="$attrs"><slot /></button>'},
				Icon: true,
			},
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
		attachTo: document.body,
	})
	await flushPromises()
	return {wrapper, router, errors}
}

function errorMessages(errors: unknown[]) {
	return errors.map(e => (e instanceof Error ? e.message : String(e)))
}

describe('ProjectSettingsArchive', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		setActivePinia(createPinia())
		document.body.innerHTML = ''
		Object.keys(projects).forEach(id => delete projects[Number(id)])
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		document.body.innerHTML = ''
	})

	it('renders without the project being loaded', async () => {
		const mounted = await mountModal('/projects/42/settings/archive')
		wrapper = mounted.wrapper

		expect(errorMessages(mounted.errors)).toEqual([])
		expect(document.querySelector('dialog.modal-dialog')?.textContent).toContain('Archive this project')
	})

	// Re-rendering with the projectId gone used to throw, which leaves Vue with a
	// half-patched subTree and breaks every later navigation (FRONTEND-OSS-2DJ).
	it('does not throw when the route changes away while still mounted', async () => {
		projects[42] = {id: 42, title: 'Test', is_archived: true} as ProjectResponse
		const mounted = await mountModal('/projects/42/settings/archive')
		wrapper = mounted.wrapper

		expect(document.querySelector('dialog.modal-dialog')?.textContent).toContain('Un-Archive this project')
		expect(document.title).toBe('Archive "Test" | Vikunja')

		await mounted.router.push('/tasks/by/upcoming')
		await flushPromises()

		expect(errorMessages(mounted.errors)).toEqual([])
		expect(document.querySelector('dialog.modal-dialog')?.textContent).toContain('Archive this project')
	})
})
