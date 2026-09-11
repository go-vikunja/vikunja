import {computed, defineComponent, h, nextTick, ref, toValue, type MaybeRefOrGetter} from 'vue'
import {mount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const useQuery = vi.hoisted(() => vi.fn())
const queryLayer = vi.hoisted(() => ({
	createSavedFilter: vi.fn(),
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
		is_favorite: false,
	})),
	deleteSavedFilter: vi.fn(),
	savedFilterQuery: vi.fn((id: number) => ({queryKey: ['saved-filters', 'detail', id]})),
	updateSavedFilter: vi.fn(),
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

import {useSavedFilter} from './useSavedFilter'

function savedFilterResponse(id: number, title: string) {
	return {
		id,
		title,
		description: '',
		filters: queryLayer.newSavedFilterDraft().filters,
		is_favorite: false,
	}
}

function mountSavedFilter(projectId?: MaybeRefOrGetter<number | undefined>) {
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
		queryLayer.createSavedFilter.mockReset()
		queryLayer.updateSavedFilter.mockReset()
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

		data.value = savedFilterResponse(1, 'Recovered')
		error.value = null
		await nextTick()

		expect(state.value.filter.value.title).toBe('Recovered')
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

	it('creates a filter from the draft and stores a clone of the response', async () => {
		const created = savedFilterResponse(7, 'New')
		queryLayer.createSavedFilter.mockResolvedValue(created)

		const {wrapper, state} = mountSavedFilter()
		state.value.filter.value.title = 'New'

		const result = await state.value.submit()

		expect(result).toBe(created)
		expect(queryLayer.createSavedFilter).toHaveBeenCalledWith({
			title: 'New',
			description: '',
			filters: queryLayer.newSavedFilterDraft().filters,
			is_favorite: false,
		})

		state.value.filter.value.filters.sort_by.push('title')
		expect(created.filters.sort_by).toEqual(['done', 'id'])
		wrapper.unmount()
	})

	it('updates the seeded draft and re-seeds it from the response', async () => {
		useQuery.mockReturnValue({
			data: ref(savedFilterResponse(1, 'Loaded')),
			isPending: ref(false),
			error: ref(null),
		})
		const updated = savedFilterResponse(1, 'Renamed')
		queryLayer.updateSavedFilter.mockResolvedValue(updated)

		const {wrapper, state} = mountSavedFilter(-2)
		state.value.filter.value.title = 'Renamed'

		const result = await state.value.submit()

		expect(result).toBe(updated)
		expect(queryLayer.createSavedFilter).not.toHaveBeenCalled()
		expect(queryLayer.updateSavedFilter).toHaveBeenCalledWith({
			id: 1,
			title: 'Renamed',
			description: '',
			filters: queryLayer.newSavedFilterDraft().filters,
			is_favorite: false,
		})
		expect(state.value.filter.value).toEqual(updated)
		expect(state.value.filter.value).not.toBe(updated)
		expect(state.value.filter.value.filters).not.toBe(updated.filters)
		wrapper.unmount()
	})

	it('does not create a filter when submitting an unloaded draft for an existing one', async () => {
		const {wrapper, state} = mountSavedFilter(-2)
		state.value.filter.value.title = 'Loaded'

		expect(await state.value.submit()).toBeUndefined()
		expect(queryLayer.createSavedFilter).not.toHaveBeenCalled()
		expect(queryLayer.updateSavedFilter).not.toHaveBeenCalled()
		wrapper.unmount()
	})

	it('only reports an invalid title once the field was touched', () => {
		const {wrapper, state} = mountSavedFilter()
		expect(state.value.titleValid.value).toBe(true)

		state.value.markTitleTouched()
		expect(state.value.titleValid.value).toBe(false)

		state.value.filter.value.title = 'Some title'
		expect(state.value.titleValid.value).toBe(true)
		wrapper.unmount()
	})
})
