import {defineComponent, ref} from 'vue'
import {flushPromises, mount, shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'
import {VueQueryPlugin} from '@tanstack/vue-query'

import XButton from '@/components/input/Button.vue'
import {newSavedFilterDraft, type SavedFilterResponse} from '@/client/queries/savedFilters'
import {i18n} from '@/i18n'

const savedFilter = vi.hoisted(() => ({
	validate: vi.fn(),
	update: vi.fn(),
}))
const router = vi.hoisted(() => ({back: vi.fn(), push: vi.fn()}))
const messages = vi.hoisted(() => ({success: vi.fn(), error: vi.fn()}))

vi.mock('@/composables/useSavedFilter', () => ({
	useSavedFilter: () => ({
		filter: ref({id: 1, ...newSavedFilterDraft(), title: 'Filter'}),
		isLoading: ref(false),
		error: ref(null),
		titleValid: ref(true),
		markTitleTouched: vi.fn(),
		validate: savedFilter.validate,
	}),
}))
vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	filtersUpdate: async (input: unknown) => ({data: await savedFilter.update(input)}),
}))
vi.mock('@/message', () => messages)
// Stubbing the async editor still fires its loader; TipTap then imports after teardown.
vi.mock('@/components/input/AsyncEditor', () => ({default: {name: 'Editor', render: () => null}}))
vi.mock('vue-router', async importOriginal => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => router,
}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

const SlotStub = defineComponent({template: '<div><slot /></div>'})
const CardStub = defineComponent({template: '<div><slot /><slot name="footer" /></div>'})

function deferred<T>() {
	let resolve!: (value: T) => void
	const promise = new Promise<T>(resolvePromise => {
		resolve = resolvePromise
	})
	return {promise, resolve}
}

function filterResponse(id: number): SavedFilterResponse {
	return {id, ...newSavedFilterDraft(), title: 'Filter', is_favorite: false}
}

const globalOptions = {
	plugins: [VueQueryPlugin],
	mocks: {$t: (key: string) => key, $router: router},
	directives: {focus: () => {}, tooltip: () => {}},
}

import FilterEdit from './FilterEdit.vue'

describe('FilterEdit', () => {
	beforeEach(() => {
		savedFilter.validate.mockReset().mockReturnValue({...newSavedFilterDraft(), title: 'Filter', is_favorite: false})
		savedFilter.update.mockReset()
		router.back.mockReset()
		messages.success.mockReset()
	})

	it('releases the submit button when saving early-returns', async () => {
		savedFilter.validate.mockReturnValue(undefined)
		const wrapper = mount(FilterEdit, {
			props: {projectId: -2},
			global: {
				...globalOptions,
				components: {Modal: SlotStub, Card: CardStub, XButton},
				stubs: {Editor: true, Filters: true, FormField: SlotStub},
			},
		})

		const primary = wrapper.findAllComponents(XButton)
			.find(button => button.props('variant') === 'primary')!
		await primary.get('button').trigger('click')
		await flushPromises()

		expect(savedFilter.validate).toHaveBeenCalledOnce()
		expect(savedFilter.update).not.toHaveBeenCalled()
		expect(primary.get('button').attributes('disabled')).toBeUndefined()
		expect(primary.classes()).not.toContain('is-loading')
		wrapper.unmount()
	})

	it('does not report success when the route switched to another filter mid-save', async () => {
		const save = deferred<SavedFilterResponse>()
		savedFilter.update.mockReturnValue(save.promise)
		const wrapper = shallowMount(FilterEdit, {
			props: {projectId: -2},
			global: {...globalOptions, stubs: {CreateEdit: SlotStub, FormField: SlotStub}},
		})

		const saving = (wrapper.vm as unknown as {save: () => Promise<void>}).save()
		await vi.waitFor(() => expect(savedFilter.update).toHaveBeenCalledOnce())
		await wrapper.setProps({projectId: -3})
		save.resolve(filterResponse(1))
		await saving

		expect(messages.success).not.toHaveBeenCalled()
		expect(router.back).not.toHaveBeenCalled()
		wrapper.unmount()
	})

	it('reports success and navigates back after saving', async () => {
		savedFilter.update.mockResolvedValue(filterResponse(1))
		const wrapper = shallowMount(FilterEdit, {
			props: {projectId: -2},
			global: {...globalOptions, stubs: {CreateEdit: SlotStub, FormField: SlotStub}},
		})

		await (wrapper.vm as unknown as {save: () => Promise<void>}).save()

		expect(messages.success).toHaveBeenCalledWith({message: i18n.global.t('filters.edit.success')})
		expect(router.back).toHaveBeenCalledOnce()
		wrapper.unmount()
	})
})
