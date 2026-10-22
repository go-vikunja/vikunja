import {computed, defineComponent, h, nextTick, ref, toValue, type MaybeRefOrGetter} from 'vue'
import {mount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const useQuery = vi.hoisted(() => vi.fn())
const queryLayer = vi.hoisted(() => ({
	newSavedFilterDraft: vi.fn(() => ({
		title: '',
		description: '',
		filters: {
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
		},
	})),
	savedFilterQuery: vi.fn((id: number) => ({queryKey: ['saved-filters', 'detail', id]})),
}))
const projectsLayer = vi.hoisted(() => ({
	getProjectIdFromSavedFilterId: vi.fn((id: number) => id > 0 ? -id - 1 : 0),
	getSavedFilterIdFromProjectId: vi.fn((id: number) => id < -1 ? -id - 1 : 0),
}))

vi.mock('@tanstack/vue-query', async importOriginal => ({
	...await importOriginal<typeof import('@tanstack/vue-query')>(),
	useQuery,
}))
vi.mock('@/client/queries/savedFilters', () => queryLayer)
vi.mock('@/client/queries/projects', () => projectsLayer)

import {useSavedFilter, useSavedFilterDraft} from './useSavedFilter'

function savedFilterResponse(id: number, title: string) {
	return {
		id,
		title,
		description: '',
		filters: queryLayer.newSavedFilterDraft().filters,
		is_favorite: false,
	}
}

function mountSavedFilter(projectId: MaybeRefOrGetter<number>) {
	let state: ReturnType<typeof useSavedFilter> | undefined
	const component = defineComponent({
		setup() {
			state = useSavedFilter(projectId)
			return () => h('div')
		},
	})
	const wrapper = mount(component)
	return {wrapper, state: computed(() => state!)}
}

describe('useSavedFilter', () => {
	beforeEach(() => {
		useQuery.mockReset()
		useQuery.mockReturnValue({
			data: ref(undefined),
			isPending: ref(true),
			error: ref(null),
		})
		queryLayer.savedFilterQuery.mockClear()
	})

	it('seeds cached data immediately and does not overwrite an edited draft later', async () => {
		const data = ref(savedFilterResponse(1, 'Cached'))
		useQuery.mockReturnValue({
			data,
			isPending: ref(false),
			error: ref(null),
		})

		const {wrapper, state} = mountSavedFilter(-2)
		expect(state.value.filter.value.title).toBe('Cached')
		expect(state.value.isLoading.value).toBe(false)
		expect(state.value.isLoaded.value).toBe(true)

		state.value.filter.value.title = 'Local edit'
		data.value = {...data.value, title: 'Background refresh'}
		await nextTick()

		expect(state.value.filter.value.title).toBe('Local edit')
		wrapper.unmount()
	})

	it('never seeds data belonging to another saved filter', () => {
		useQuery.mockReturnValue({
			data: ref(savedFilterResponse(1, 'Other filter')),
			isPending: ref(false),
			error: ref(null),
		})

		const {wrapper, state} = mountSavedFilter(-43)
		expect(state.value.filter.value.id).toBe(0)
		expect(state.value.filter.value.title).toBe('')
		expect(state.value.isLoaded.value).toBe(false)
		wrapper.unmount()
	})

	it('settles the loading state when the detail query fails', async () => {
		const data = ref<ReturnType<typeof savedFilterResponse> | undefined>(undefined)
		const error = ref<Error | null>(null)
		const isPending = ref(true)
		useQuery.mockReturnValue({
			data,
			isPending,
			error,
		})

		const {wrapper, state} = mountSavedFilter(-2)
		expect(state.value.isLoading.value).toBe(true)

		error.value = new Error('Not found')
		isPending.value = false
		await nextTick()

		expect(state.value.isLoading.value).toBe(false)
		expect(state.value.error.value).toBe(error.value)
		expect(state.value.isLoaded.value).toBe(false)

		data.value = savedFilterResponse(1, 'Recovered')
		error.value = null
		await nextTick()

		expect(state.value.filter.value.title).toBe('Recovered')
		expect(state.value.isLoaded.value).toBe(true)
		wrapper.unmount()
	})

	it('re-seeds the form when returning to a saved filter the instance already showed', async () => {
		const cached = savedFilterResponse(1, 'Cached')
		useQuery.mockImplementation(options => {
			const savedFilterId = computed(() => toValue(options).queryKey[2] as number)
			return {
				data: computed(() => savedFilterId.value > 0 ? cached : undefined),
				isPending: computed(() => savedFilterId.value <= 0),
				error: ref(null),
			}
		})

		const projectId = ref<number>(-2)
		const {wrapper, state} = mountSavedFilter(projectId)
		expect(state.value.filter.value.title).toBe('Cached')

		projectId.value = 5
		await nextTick()
		expect(state.value.filter.value.id).toBe(0)
		expect(state.value.filter.value.title).toBe('')

		projectId.value = -2
		await nextTick()
		expect(state.value.filter.value.id).toBe(1)
		expect(state.value.filter.value.title).toBe('Cached')
		wrapper.unmount()
	})

	it('validates a new local draft without subscribing to a query', () => {
		const draft = useSavedFilterDraft()
		draft.filter.value.title = 'New'

		const result = draft.validate()

		expect(useQuery).not.toHaveBeenCalled()
		expect(result).toEqual({
			title: 'New',
			description: '',
			filters: queryLayer.newSavedFilterDraft().filters,
		})

		draft.filter.value.filters.sort_by.push('title')
		expect(result?.filters.sort_by).toEqual(['done', 'id'])
	})

	it('returns a writable copy without mutating the cached filter', () => {
		const cached = savedFilterResponse(1, 'Loaded')
		useQuery.mockReturnValue({
			data: ref(cached),
			isPending: ref(false),
			error: ref(null),
		})
		const {wrapper, state} = mountSavedFilter(-2)
		state.value.filter.value.title = 'Renamed'

		const result = state.value.validate()

		expect(result).toEqual({
			title: 'Renamed',
			description: '',
			filters: queryLayer.newSavedFilterDraft().filters,
			is_favorite: false,
		})
		state.value.filter.value.filters.sort_by.push('title')
		state.value.filter.value.filters.order_by.push('asc')
		expect(cached.title).toBe('Loaded')
		expect(cached.filters.sort_by).toEqual(['done', 'id'])
		expect(cached.filters.order_by).toEqual(['asc', 'desc'])
		expect(result?.filters.sort_by).toEqual(['done', 'id'])
		wrapper.unmount()
	})

	it('submits the cached favorite flag instead of a form value', () => {
		const cached = ref(savedFilterResponse(1, 'Loaded'))
		useQuery.mockReturnValue({
			data: cached,
			isPending: ref(false),
			error: ref(null),
		})
		const {wrapper, state} = mountSavedFilter(-2)
		cached.value.is_favorite = true

		expect(state.value.validate()).toMatchObject({title: 'Loaded', is_favorite: true})
		expect(state.value.filter.value).not.toHaveProperty('is_favorite')
		wrapper.unmount()
	})

	it('does not validate an unloaded draft for an existing filter', () => {
		const {wrapper, state} = mountSavedFilter(-2)
		state.value.filter.value.title = 'Loaded'

		expect(state.value.validate()).toBeUndefined()
		expect(state.value.isLoaded.value).toBe(false)
		wrapper.unmount()
	})

	it('only reports an invalid title once the field was touched', () => {
		const draft = useSavedFilterDraft()
		expect(draft.titleValid.value).toBe(true)

		draft.markTitleTouched()
		expect(draft.titleValid.value).toBe(false)

		draft.filter.value.title = 'Some title'
		expect(draft.titleValid.value).toBe(true)
	})

	it('validates the title on submit without returning an empty-title payload', () => {
		const draft = useSavedFilterDraft()

		expect(draft.validate()).toBeUndefined()
		expect(draft.titleValid.value).toBe(false)
	})
})
