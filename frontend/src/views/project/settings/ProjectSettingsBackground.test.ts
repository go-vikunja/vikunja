import {describe, it, expect, beforeEach, vi} from 'vitest'
import {defineComponent, h} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'

import ProjectSettingsBackground from '@/views/project/settings/ProjectSettingsBackground.vue'
import {useBaseStore} from '@/stores/base'
import {useConfigStore} from '@/stores/config'
import {getBlobFromBlurHash} from '@/helpers/getBlobFromBlurHash'
import type {IProject} from '@/modelTypes/IProject'
import en from '@/i18n/lang/en.json'

const REMOVE_BACKGROUND_LABEL = 'Remove Background'

const unsplashResults = [{
	id: 'img-1',
	blurHash: 'LEHV6nWB2yk8pyo0adR*.7kCMdnj',
	info: {author: 'someone', authorName: 'Some One'},
}]

const getAll = vi.fn(async () => unsplashResults.slice())
const thumb = vi.fn(async () => 'thumb-url')

vi.mock('@/services/backgroundUnsplash', () => ({
	default: class {
		loading = false
		getAll = getAll
		thumb = thumb
		update = vi.fn()
	},
}))

vi.mock('@/services/backgroundUpload', () => ({
	default: class {
		loading = false
		create = vi.fn()
	},
}))

vi.mock('@/services/project', () => ({
	default: class {
		loading = false
		removeBackground = vi.fn()
	},
}))

vi.mock('@/helpers/getBlobFromBlurHash', () => ({
	getBlobFromBlurHash: vi.fn(async () => null),
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

async function mountPage({
	providers = ['upload'],
	project = null,
}: {
	providers?: ('unsplash' | 'upload')[]
	project?: IProject | null
} = {}) {
	const errors: unknown[] = []
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{
			path: '/projects/:projectId/settings/background',
			name: 'project.settings.background',
			component: ProjectSettingsBackground,
		}],
	})
	await router.push('/projects/1/settings/background')
	await router.isReady()

	// The stores can only be created from within a component setup because the
	// base store itself calls useI18n().
	const Host = defineComponent({
		setup() {
			useConfigStore().enabledBackgroundProviders = providers
			useBaseStore().setCurrentProject(project)
			return () => h(ProjectSettingsBackground)
		},
	})

	const wrapper = mount(Host, {
		global: {
			plugins: [i18n, router],
			stubs: {
				CreateEdit: {template: '<div><slot /><slot name="footer" /></div>'},
				XButton: {template: '<button type="button"><slot /></button>'},
				BaseButton: {template: '<a><slot /></a>'},
				CustomTransition: {template: '<div><slot /></div>'},
			},
			config: {
				errorHandler(err: unknown) {
					errors.push(err)
				},
			},
		},
	})
	await flushPromises()
	return {wrapper, errors}
}

describe('ProjectSettingsBackground', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		vi.clearAllMocks()
		window.URL.createObjectURL = vi.fn(() => 'blob:fake')
	})

	it('renders without the remove button when no project is loaded', async () => {
		const {wrapper, errors} = await mountPage({project: null})

		expect(errors).toEqual([])
		expect(wrapper.text()).not.toContain(REMOVE_BACKGROUND_LABEL)
	})

	it('shows the remove button when the current project has a background', async () => {
		const {wrapper} = await mountPage({
			project: {id: 1, backgroundInformation: {type: 'upload'}} as unknown as IProject,
		})

		expect(wrapper.text()).toContain(REMOVE_BACKGROUND_LABEL)
	})

	it('does not create an object url when the blur hash could not be decoded', async () => {
		const {errors} = await mountPage({providers: ['unsplash']})

		expect(getBlobFromBlurHash).toHaveBeenCalled()
		expect(window.URL.createObjectURL).not.toHaveBeenCalled()
		expect(errors).toEqual([])
	})

	it('creates an object url for a decoded blur hash', async () => {
		vi.mocked(getBlobFromBlurHash).mockResolvedValueOnce(new Blob(['x']))

		await mountPage({providers: ['unsplash']})

		expect(window.URL.createObjectURL).toHaveBeenCalledOnce()
	})
})
