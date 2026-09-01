import {defineComponent, ref, toValue, type Ref} from 'vue'
import {flushPromises, shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {Image} from '@/client/generated'
import type {ProjectResponse} from '@/client/queries/projects'

type SearchData = {pages: Image[][]}

const state = vi.hoisted(() => ({
	currentProjectId: 0,
	routeParams: undefined as {projectId: string} | undefined,
	project: undefined as Ref<ProjectResponse | undefined> | undefined,
	projectIsError: undefined as Ref<boolean> | undefined,
	searchData: undefined as Ref<SearchData | undefined> | undefined,
	searchIsError: undefined as Ref<boolean> | undefined,
}))

const refetchSearch = vi.hoisted(() => vi.fn())
const infiniteQueryOptions = vi.hoisted(() => [] as unknown[])

const handleSetCurrentProject = vi.hoisted(() => vi.fn())
const routerBack = vi.hoisted(() => vi.fn())
const success = vi.hoisted(() => vi.fn())
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
		unsplashAuthor: () => null,
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
		useQuery: () => ({
			data: state.project,
			error: ref(null),
			isPending: ref(false),
			isError: state.projectIsError,
		}),
		useInfiniteQuery: (options: unknown) => {
			infiniteQueryOptions.push(options)
			return {
				data: state.searchData,
				isFetching: ref(false),
				hasNextPage: ref(false),
				isFetchingNextPage: ref(false),
				fetchNextPage: vi.fn(),
				isError: state.searchIsError,
				refetch: refetchSearch,
			}
		},
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
		state.currentProjectId = 7
		state.routeParams!.projectId = '7'
		state.project = ref(project())
		state.projectIsError = ref(false)
		state.searchData = ref({pages: []})
		state.searchIsError = ref(false)
		infiniteQueryOptions.splice(0)
		handleSetCurrentProject.mockClear()
		routerBack.mockClear()
		refetchSearch.mockClear()
		success.mockClear()
		backgrounds.deleteProjectBackground.mockReset()
		backgrounds.setUnsplashProjectBackground.mockReset()
		backgrounds.uploadProjectBackground.mockReset()
	})

	it('applies the removed background to the base store and navigates back', async () => {
		const updated = project({background_information: null, background_blur_hash: ''})
		backgrounds.deleteProjectBackground.mockResolvedValue(updated)
		const wrapper = mountView()

		await clickRemove(wrapper)

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(handleSetCurrentProject).toHaveBeenCalledWith({project: updated, forceUpdate: true})
		expect(success).toHaveBeenCalled()
		expect(routerBack).toHaveBeenCalled()
	})

	it('still reports success when another project is currently displayed', async () => {
		state.currentProjectId = 99
		backgrounds.deleteProjectBackground.mockResolvedValue(project({background_information: null}))
		const wrapper = mountView()

		await clickRemove(wrapper)

		expect(backgrounds.deleteProjectBackground).toHaveBeenCalledWith(7)
		expect(handleSetCurrentProject).not.toHaveBeenCalled()
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
		expect(handleSetCurrentProject).not.toHaveBeenCalled()
		expect(success).not.toHaveBeenCalled()
		expect(routerBack).not.toHaveBeenCalled()
	})

	it.each([0, null])('hides every background action for max_permission %s', maxPermission => {
		state.project = ref(project({max_permission: maxPermission} as Partial<ProjectResponse>))
		const wrapper = mountView()

		expect(findButton(wrapper, 'project.background.upload')).toBeUndefined()
		expect(findButton(wrapper, 'project.background.remove')).toBeUndefined()
		expect(wrapper.find('input[type="text"]').exists()).toBe(false)
		expect(toValue(infiniteQueryOptions[0])).toMatchObject({enabled: false})
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
