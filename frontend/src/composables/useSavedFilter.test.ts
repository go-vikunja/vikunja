import {computed, defineComponent, h, nextTick, ref, toValue} from 'vue'
import {mount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const useQuery = vi.hoisted(() => vi.fn())
const queryLayer = vi.hoisted(() => ({
	createSavedFilter: vi.fn(),
	createSavedFilterDraft: vi.fn(() => ({
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
	getProjectIdFromSavedFilterId: vi.fn((id: number) => id > 0 ? -id - 1 : 0),
	getSavedFilterIdFromProjectId: vi.fn((id: number) => id < -1 ? -id - 1 : 0),
	savedFilterQuery: vi.fn((id: number) => ({queryKey: ['saved-filters', 'detail', id]})),
	updateSavedFilter: vi.fn(),
}))

vi.mock('@tanstack/vue-query', async importOriginal => ({
	...await importOriginal<typeof import('@tanstack/vue-query')>(),
	useQuery,
}))
vi.mock('@/client/queries/savedFilters', () => queryLayer)
vi.mock('vue-router', () => ({
	useRouter: () => ({push: vi.fn(), back: vi.fn()}),
}))
vi.mock('vue-i18n', () => ({
	useI18n: () => ({t: (key: string) => key}),
}))
vi.mock('@/message', () => ({success: vi.fn()}))

import {useSavedFilter} from './useSavedFilter'

function mountSavedFilter(projectId?: number) {
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
			isFetching: ref(true),
			isError: ref(false),
			error: ref(null),
		})
		queryLayer.savedFilterQuery.mockClear()
	})

	it('keeps a new filter editable while the disabled detail query is pending', () => {
		const {wrapper, state} = mountSavedFilter()

		expect(state.value.isLoading.value).toBe(false)
		expect(toValue(useQuery.mock.calls[0][0])).toMatchObject({enabled: false})
		wrapper.unmount()
	})

	it('loads the saved-filter detail derived from a pseudo-project id', () => {
		const {wrapper, state} = mountSavedFilter(-43)
		const options = toValue(useQuery.mock.calls[0][0])

		expect(queryLayer.savedFilterQuery).toHaveBeenCalledWith(42)
		expect(options).toMatchObject({enabled: true})
		expect(state.value.isLoading.value).toBe(true)
		wrapper.unmount()
	})

	it('waits for the initial refresh and does not overwrite an edited draft later', async () => {
		const data = ref({
			id: 1,
			title: 'Cached',
			description: '',
			filters: queryLayer.createSavedFilterDraft().filters,
			is_favorite: false,
		})
		const isFetching = ref(true)
		useQuery.mockReturnValue({
			data,
			isPending: ref(false),
			isFetching,
			isError: ref(false),
			error: ref(null),
		})

		const {wrapper, state} = mountSavedFilter(-2)
		expect(state.value.filter.value.title).toBe('')
		expect(state.value.isLoading.value).toBe(true)

		data.value = {...data.value, title: 'Fresh'}
		isFetching.value = false
		await nextTick()
		expect(state.value.filter.value.title).toBe('Fresh')

		state.value.filter.value.title = 'Local edit'
		isFetching.value = true
		data.value = {...data.value, title: 'Background refresh'}
		isFetching.value = false
		await nextTick()

		expect(state.value.filter.value.title).toBe('Local edit')
		wrapper.unmount()
	})

	it('settles the loading state when the detail query fails', async () => {
		const data = ref<{
			id: number
			title: string
			description: string
			filters: ReturnType<typeof queryLayer.createSavedFilterDraft>['filters']
			is_favorite: boolean
		} | undefined>(undefined)
		const error = ref<Error | null>(null)
		const isError = ref(false)
		const isFetching = ref(true)
		useQuery.mockReturnValue({
			data,
			isPending: ref(false),
			isFetching,
			isError,
			error,
		})

		const {wrapper, state} = mountSavedFilter(-2)
		expect(state.value.isLoading.value).toBe(true)

		error.value = new Error('Not found')
		isError.value = true
		isFetching.value = false
		await nextTick()

		expect(state.value.isLoading.value).toBe(false)
		expect(state.value.error.value).toBe(error.value)

		data.value = {
			id: 1,
			title: 'Recovered',
			description: '',
			filters: queryLayer.createSavedFilterDraft().filters,
			is_favorite: false,
		}
		error.value = null
		isError.value = false
		await nextTick()

		expect(state.value.filter.value.title).toBe('Recovered')
		wrapper.unmount()
	})
})
