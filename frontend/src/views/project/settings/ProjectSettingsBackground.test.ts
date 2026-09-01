import {defineComponent, ref, type Ref} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {ProjectResponse} from '@/client/queries/projects'
import type {UnsplashSearchImage} from '@/client/queries/projectBackgrounds'

type SearchData = {pages: UnsplashSearchImage[][]}

const state = vi.hoisted(() => ({
	currentProjectId: 0,
	project: undefined as Ref<ProjectResponse | undefined> | undefined,
	searchData: undefined as Ref<SearchData | undefined> | undefined,
}))

const handleSetCurrentProject = vi.hoisted(() => vi.fn())
const routerBack = vi.hoisted(() => vi.fn())
const backgrounds = vi.hoisted(() => ({
	deleteProjectBackground: vi.fn(),
	setUnsplashProjectBackground: vi.fn(),
	uploadProjectBackground: vi.fn(),
}))

vi.mock('@/client/queries/projectBackgrounds', async () => {
	const {ref} = await import('vue')
	return {
		...backgrounds,
		projectBackgroundQuery: (projectId: number) => ({queryKey: ['project-backgrounds', 'project', projectId]}),
		unsplashBackgroundSearchQuery: (query: string) => ({queryKey: ['project-backgrounds', 'unsplash', 'search', query]}),
		unsplashBackgroundThumbnailQuery: (imageId: string) => ({queryKey: ['project-backgrounds', 'unsplash', 'thumbnail', imageId]}),
		useDeleteProjectBackgroundMutation: () => ({
			isPending: ref(false),
			mutateAsync: backgrounds.deleteProjectBackground,
		}),
		useSetUnsplashProjectBackgroundMutation: () => ({
			isPending: ref(false),
			mutateAsync: backgrounds.setUnsplashProjectBackground,
		}),
		useUploadProjectBackgroundMutation: () => ({
			isPending: ref(false),
			mutateAsync: backgrounds.uploadProjectBackground,
		}),
	}
})

vi.mock('@tanstack/vue-query', async importOriginal => {
	const {ref} = await import('vue')
	return {
		...await importOriginal<typeof import('@tanstack/vue-query')>(),
		useQuery: () => ({data: state.project, error: ref(null), isPending: ref(false)}),
		useInfiniteQuery: () => ({
			data: state.searchData,
			isFetching: ref(false),
			hasNextPage: ref(false),
			isFetchingNextPage: ref(false),
			fetchNextPage: vi.fn(),
		}),
	}
})

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({
		get currentProjectId() {
			return state.currentProjectId
		},
		handleSetCurrentProject,
	}),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({enabledBackgroundProviders: ['unsplash', 'upload']}),
}))

vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))
vi.mock('vue-router', () => ({
	useRoute: () => ({params: {projectId: '7'}}),
	useRouter: () => ({back: routerBack}),
}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectSettingsBackground from './ProjectSettingsBackground.vue'

const SlotStub = defineComponent({template: '<div><slot /></div>'})
const BaseButtonStub = defineComponent({template: '<a><slot /></a>'})

function project(overrides: Partial<ProjectResponse> = {}) {
	return {
		id: 7,
		title: 'Project',
		max_permission: 1,
		background_information: {provider: 'upload'},
		background_blur_hash: 'hash',
		...overrides,
	} as ProjectResponse
}

function mountView() {
	return shallowMount(ProjectSettingsBackground, {
		global: {
			stubs: {BaseButton: BaseButtonStub, CreateEdit: SlotStub},
			mocks: {$t: (key: string) => key, $router: {back: routerBack}},
			directives: {focus: () => {}, tooltip: () => {}},
		},
	})
}

type Vm = {removeBackground: () => Promise<void>}

describe('ProjectSettingsBackground', () => {
	beforeEach(() => {
		state.currentProjectId = 7
		state.project = ref(project())
		state.searchData = ref({pages: []})
		handleSetCurrentProject.mockClear()
		routerBack.mockClear()
		backgrounds.deleteProjectBackground.mockReset()
		backgrounds.setUnsplashProjectBackground.mockReset()
		backgrounds.uploadProjectBackground.mockReset()
	})

	it('applies the removed background to the base store and navigates back', async () => {
		const updated = project({background_information: null, background_blur_hash: ''})
		backgrounds.deleteProjectBackground.mockResolvedValue(updated)
		const wrapper = mountView()

		await (wrapper.vm as unknown as Vm).removeBackground()

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(handleSetCurrentProject).toHaveBeenCalledWith({project: updated, forceUpdate: true})
		expect(routerBack).toHaveBeenCalled()
	})

	it('does not touch the base store when the current project changed while removing', async () => {
		backgrounds.deleteProjectBackground.mockImplementation(async () => {
			state.currentProjectId = 99
			return project({background_information: null})
		})
		const wrapper = mountView()

		await (wrapper.vm as unknown as Vm).removeBackground()

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(handleSetCurrentProject).not.toHaveBeenCalled()
		expect(routerBack).not.toHaveBeenCalled()
	})

	it('renders an encoded attribution link only for images with a known author', () => {
		state.searchData = ref({pages: [[
			{id: 'no-author', blur_hash: 'hash', author: '', author_name: ''},
			{id: 'with-author', blur_hash: 'hash', author: 'a b', author_name: 'A B'},
		]]})
		const wrapper = mountView()

		const links = wrapper.findAll('.image-search__info')

		expect(links).toHaveLength(1)
		expect(links[0].attributes('href')).toBe('https://unsplash.com/@a%20b?utm_source=vikunja&utm_medium=referral')
		expect(links[0].text()).toBe('A B')
	})
})
