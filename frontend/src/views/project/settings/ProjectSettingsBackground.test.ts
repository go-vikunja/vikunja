import {defineComponent, ref, type Ref} from 'vue'
import {flushPromises, shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {VueQueryPlugin} from '@tanstack/vue-query'
import {queryClient} from '@/client/queryClient'
import {projectKeys} from '@/client/queries/projects'

import type {Image} from '@/client/generated'
import type {ProjectResponse} from '@/client/queries/projects'

type SearchData = {pages: Image[][]}

const state = vi.hoisted(() => ({
	routeParams: undefined as {projectId: string} | undefined,
	project: undefined as Ref<ProjectResponse | undefined> | undefined,
	projectIsError: undefined as Ref<boolean> | undefined,
	searchData: undefined as Ref<SearchData | undefined> | undefined,
	searchIsError: undefined as Ref<boolean> | undefined,
}))

const refetchSearch = vi.hoisted(() => vi.fn())

const routerBack = vi.hoisted(() => vi.fn())
const success = vi.hoisted(() => vi.fn())
const backgrounds = vi.hoisted(() => ({
	deleteProjectBackground: vi.fn(),
	setUnsplashProjectBackground: vi.fn(),
	uploadProjectBackground: vi.fn(),
}))

vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	projectsBackgroundDelete: async ({path}: {path: {project: number}}) => ({
		data: await backgrounds.deleteProjectBackground(path.project),
	}),
	projectsBackgroundUnsplashSet: async (input: unknown) => ({data: await backgrounds.setUnsplashProjectBackground(input)}),
	projectsBackgroundUpload: async (input: unknown) => ({data: await backgrounds.uploadProjectBackground(input)}),
}))

vi.mock('@tanstack/vue-query', async importOriginal => {
	const {ref} = await import('vue')
	return {
		...await importOriginal<typeof import('@tanstack/vue-query')>(),
		useQuery: () => ({
			data: state.project,
			error: ref(null),
			isPending: ref(false),
			isError: state.projectIsError,
		}),
		useInfiniteQuery: () => ({
			data: state.searchData,
			isFetching: ref(false),
			hasNextPage: ref(false),
			isFetchingNextPage: ref(false),
			fetchNextPage: vi.fn(),
			isError: state.searchIsError,
			refetch: refetchSearch,
		}),
	}
})

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({enabled_background_providers: ['unsplash', 'upload']}),
}))

vi.mock('@/message', () => ({error: vi.fn(), success}))
vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))
vi.mock('vue-router', async () => {
	const {reactive} = await import('vue')
	state.routeParams = reactive({projectId: '7'})
	return {
		useRoute: () => ({params: state.routeParams}),
		useRouter: () => ({back: routerBack}),
	}
})
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import UnsplashBackgroundThumbnail from '@/components/project/partials/UnsplashBackgroundThumbnail.vue'
import ProjectSettingsBackground from './ProjectSettingsBackground.vue'

const CreateEditStub = defineComponent({template: '<div><slot /><slot name="footer" /></div>'})
const BaseButtonStub = defineComponent({template: '<a><slot /></a>'})
const XButtonStub = defineComponent({emits: ['click'], template: '<button @click="$emit(\'click\', $event)"><slot /></button>'})

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
			plugins: [[VueQueryPlugin, {queryClient}]],
			stubs: {BaseButton: BaseButtonStub, CreateEdit: CreateEditStub, XButton: XButtonStub},
			mocks: {$t: (key: string) => key, $router: {back: routerBack}},
			directives: {focus: () => {}, tooltip: () => {}},
		},
	})
}

function findButton(wrapper: ReturnType<typeof mountView>, text: string) {
	return wrapper.findAll('button').find(button => button.text() === text)
}

async function clickRemove(wrapper: ReturnType<typeof mountView>) {
	await findButton(wrapper, 'project.background.remove')?.trigger('click')
	await flushPromises()
}

describe('ProjectSettingsBackground', () => {
	beforeEach(() => {
		vi.restoreAllMocks()
		queryClient.clear()
		queryClient.setQueryData(projectKeys.detail(7), project())
		state.routeParams!.projectId = '7'
		state.project = ref(project())
		state.projectIsError = ref(false)
		state.searchData = ref({pages: []})
		state.searchIsError = ref(false)
		routerBack.mockClear()
		refetchSearch.mockClear()
		success.mockClear()
		backgrounds.deleteProjectBackground.mockReset()
		backgrounds.setUnsplashProjectBackground.mockReset()
		backgrounds.uploadProjectBackground.mockReset()
	})

	it('clears the cached background and navigates back after removing it', async () => {
		backgrounds.deleteProjectBackground.mockResolvedValue({id: 7})
		const wrapper = mountView()

		await clickRemove(wrapper)

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(queryClient.getQueryData<ProjectResponse>(projectKeys.detail(7)))
			.toMatchObject({background_information: null, background_blur_hash: ''})
		expect(success).toHaveBeenCalled()
		expect(routerBack).toHaveBeenCalled()
	})

	it('does nothing when the route left the project while removing', async () => {
		backgrounds.deleteProjectBackground.mockImplementation(async () => {
			state.routeParams!.projectId = '99'
			return project({background_information: null})
		})
		const wrapper = mountView()

		await clickRemove(wrapper)

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(success).not.toHaveBeenCalled()
		expect(routerBack).not.toHaveBeenCalled()
	})

	it.each([0, null])('hides every background action for max_permission %s', maxPermission => {
		state.project = ref(project({max_permission: maxPermission} as Partial<ProjectResponse>))
		const wrapper = mountView()

		expect(findButton(wrapper, 'project.background.upload')).toBeUndefined()
		expect(findButton(wrapper, 'project.background.remove')).toBeUndefined()
		expect(wrapper.find('input[type="text"]').exists()).toBe(false)
		expect(wrapper.text()).toContain('project.background.noPermission')
	})

	it('renders a thumbnail only for search results with an id', () => {
		state.searchData = ref({pages: [[{blur_hash: 'hash'}, {id: 'with-id', blur_hash: 'hash'}]]})
		const wrapper = mountView()

		const thumbnails = wrapper.findAllComponents(UnsplashBackgroundThumbnail)

		expect(thumbnails).toHaveLength(1)
		expect(thumbnails[0].props('image')).toMatchObject({id: 'with-id'})
	})

	it('shows a retry button when the unsplash search failed and refetches on click', async () => {
		state.searchIsError = ref(true)
		const wrapper = mountView()

		expect(wrapper.text()).toContain('project.background.searchError')

		await findButton(wrapper, 'project.background.retry')?.trigger('click')

		expect(refetchSearch).toHaveBeenCalled()
	})
})
