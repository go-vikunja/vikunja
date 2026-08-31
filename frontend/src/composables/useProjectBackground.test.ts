import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {computed, defineComponent, h, nextTick, ref, toValue, type Ref} from 'vue'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'

import type {ProjectResponse} from '@/client/queries/projects'

const queryLayer = vi.hoisted(() => ({
	projectBackgroundQuery: vi.fn((projectId: number) => ({
		queryKey: ['project-backgrounds', 'project', projectId],
	})),
}))
const useQuery = vi.hoisted(() => vi.fn())
const getBlobFromBlurHash = vi.hoisted(() => vi.fn())

vi.mock('@/client/queries/projectBackgrounds', () => queryLayer)
vi.mock('@tanstack/vue-query', async (importOriginal) => ({
	...await importOriginal<typeof import('@tanstack/vue-query')>(),
	useQuery,
}))
vi.mock('@/helpers/getBlobFromBlurHash', () => ({getBlobFromBlurHash}))

import {useProjectBackground} from './useProjectBackground'

function project(overrides: Partial<ProjectResponse> = {}): ProjectResponse {
	return {
		id: 7,
		title: 'Project',
		description: '',
		hex_color: '',
		identifier: '',
		is_archived: false,
		is_favorite: false,
		parent_project_id: 0,
		position: 0,
		views: [],
		background_information: {provider: 'upload'},
		background_blur_hash: 'hash-one',
		...overrides,
	}
}

type BackgroundState = ReturnType<typeof useProjectBackground>

function mountBackground(projectValue: Ref<ProjectResponse | null>) {
	let state: BackgroundState | undefined
	const component = defineComponent({
		setup() {
			state = useProjectBackground(projectValue)
			return () => h('div')
		},
	})
	const wrapper = mount(component)
	return {wrapper, state: computed(() => state!)}
}

describe('useProjectBackground', () => {
	let backgroundBlob: ReturnType<typeof ref<Blob | undefined>>
	let isPending: ReturnType<typeof ref<boolean>>
	let createObjectURL: ReturnType<typeof vi.fn<(blob: Blob) => string>>
	let revokeObjectURL: ReturnType<typeof vi.fn<(url: string) => void>>
	let wrappers: VueWrapper[]

	beforeEach(() => {
		backgroundBlob = ref<Blob>()
		isPending = ref(false)
		useQuery.mockReset()
		useQuery.mockReturnValue({data: backgroundBlob, isPending})
		queryLayer.projectBackgroundQuery.mockClear()
		getBlobFromBlurHash.mockReset()

		let objectUrl = 0
		createObjectURL = vi.fn(() => `blob:${++objectUrl}`)
		revokeObjectURL = vi.fn()
		window.URL.createObjectURL = createObjectURL
		window.URL.revokeObjectURL = revokeObjectURL
		wrappers = []
	})

	afterEach(() => {
		wrappers.splice(0).forEach(wrapper => wrapper.unmount())
	})

	it('keys the query from generated project fields and disables it without a background', async () => {
		getBlobFromBlurHash.mockResolvedValue(null)
		const currentProject = ref<ProjectResponse | null>(project())
		const mounted = mountBackground(currentProject)
		wrappers.push(mounted.wrapper)

		const options = useQuery.mock.calls[0][0]
		expect(toValue(options)).toMatchObject({
			queryKey: ['project-backgrounds', 'project', 7],
			enabled: true,
		})

		currentProject.value = project({id: 8, background_information: null})
		await nextTick()

		expect(toValue(options)).toMatchObject({
			queryKey: ['project-backgrounds', 'project', 8],
			enabled: false,
		})
	})

	it('revokes replaced and unmounted background and BlurHash URLs', async () => {
		const firstBackground = new Blob(['background-one'])
		const secondBackground = new Blob(['background-two'])
		const firstBlur = new Blob(['blur-one'])
		const secondBlur = new Blob(['blur-two'])
		backgroundBlob.value = firstBackground
		getBlobFromBlurHash
			.mockResolvedValueOnce(firstBlur)
			.mockResolvedValueOnce(secondBlur)
		const currentProject = ref<ProjectResponse | null>(project())
		const mounted = mountBackground(currentProject)
		wrappers.push(mounted.wrapper)
		await flushPromises()

		expect(mounted.state.value.background.value).toBe('blob:1')
		expect(mounted.state.value.blurHashUrl.value).toBe('blob:2')

		backgroundBlob.value = secondBackground
		await nextTick()
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:1')
		expect(mounted.state.value.background.value).toBe('blob:3')

		currentProject.value = project({background_blur_hash: 'hash-two'})
		await flushPromises()
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:2')
		expect(mounted.state.value.blurHashUrl.value).toBe('blob:4')

		mounted.wrapper.unmount()
		wrappers = []
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:3')
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:4')
	})

	it('clears owned URLs when the project loses its background', async () => {
		backgroundBlob.value = new Blob(['background'])
		getBlobFromBlurHash.mockResolvedValue(new Blob(['blur']))
		const currentProject = ref<ProjectResponse | null>(project())
		const mounted = mountBackground(currentProject)
		wrappers.push(mounted.wrapper)
		await flushPromises()

		currentProject.value = project({background_information: null})
		await nextTick()

		expect(mounted.state.value.background.value).toBeNull()
		expect(mounted.state.value.blurHashUrl.value).toBe('')
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:1')
		expect(revokeObjectURL).toHaveBeenCalledWith('blob:2')
	})

	it('does not create a BlurHash URL after unmount', async () => {
		let resolveBlur: (blob: Blob) => void = () => {}
		getBlobFromBlurHash.mockReturnValue(new Promise(resolve => {
			resolveBlur = resolve
		}))
		const currentProject = ref<ProjectResponse | null>(project())
		const mounted = mountBackground(currentProject)
		mounted.wrapper.unmount()

		resolveBlur(new Blob(['late-blur']))
		await flushPromises()

		expect(createObjectURL).not.toHaveBeenCalled()
	})
})
