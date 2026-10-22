import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {computed, defineComponent, h, nextTick, ref, type Ref} from 'vue'
import {mount, type VueWrapper} from '@vue/test-utils'

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
	let wrappers: VueWrapper[]

	beforeEach(() => {
		backgroundBlob = ref<Blob>()
		isPending = ref(false)
		useQuery.mockReset()
		useQuery.mockReturnValue({data: backgroundBlob, isPending})
		queryLayer.projectBackgroundQuery.mockClear()
		getBlobFromBlurHash.mockReset()

		let objectUrl = 0
		window.URL.createObjectURL = vi.fn(() => `blob:${++objectUrl}`)
		window.URL.revokeObjectURL = vi.fn()
		wrappers = []
	})

	afterEach(() => {
		wrappers.splice(0).forEach(wrapper => wrapper.unmount())
	})

	it('ignores a cached blob while the project has no background', async () => {
		getBlobFromBlurHash.mockResolvedValue(null)
		backgroundBlob.value = new Blob(['background'])
		const currentProject = ref<ProjectResponse | null>(project({background_information: null}))
		const mounted = mountBackground(currentProject)
		wrappers.push(mounted.wrapper)

		expect(mounted.state.value.background.value).toBeUndefined()

		currentProject.value = project()
		await nextTick()

		expect(mounted.state.value.background.value).toBe('blob:1')
	})
})
