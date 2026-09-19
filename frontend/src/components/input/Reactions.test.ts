import {describe, it, expect, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {taskKeys, normalizeTask} from '@/client/queries/tasks'

import Reactions from './Reactions.vue'
import en from '@/i18n/lang/en.json'

const CURRENT_USER = {
	id: 1,
	username: 'current',
}

const OTHER_USER = {
	id: 2,
	username: 'other',
}

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		info: CURRENT_USER,
		settings: {frontendSettings: {colorSchema: 'light'}},
	}),
}))

vi.mock('@/client/generated', () => ({
	reactionsCreate: vi.fn(async () => ({data: {}})),
	reactionsDelete: vi.fn(async () => ({data: undefined})),
}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))

vi.mock('vuemoji-picker', () => ({
	VuemojiPicker: {template: '<div />'},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountReactions(modelValue: Record<string, typeof CURRENT_USER[]>) {
	const queryClient = new QueryClient()
	queryClient.setQueryData(taskKeys.detail(1), normalizeTask({id: 1, reactions: modelValue}))
	return mount(Reactions, {
		props: {
			entityKind: 'tasks' as const,
			entityId: 1,
			modelValue,
		},
		global: {
			plugins: [i18n, [VueQueryPlugin, {queryClient}]],
			stubs: {
				BaseButton: {template: '<button><slot /></button>'},
				Icon: true,
				CustomTransition: {template: '<div><slot /></div>'},
			},
			directives: {tooltip: {}},
		},
	})
}

describe('Reactions', () => {
	it('emits a new object when adding a reaction', async () => {
		const modelValue = {'🎉': [OTHER_USER]}
		const wrapper = mountReactions(modelValue)

		await wrapper.findAll('button')[0].trigger('click')
		await flushPromises()

		const emitted = wrapper.emitted('update:modelValue')
		expect(emitted).toHaveLength(1)
		expect(emitted![0][0]).toEqual({'🎉': [OTHER_USER, CURRENT_USER]})
		expect(emitted![0][0]).not.toBe(modelValue)
		expect(modelValue).toEqual({'🎉': [OTHER_USER]})
		expect(modelValue['🎉']).toHaveLength(1)
	})

	it('emits a new object without the emoji when removing the last reaction', async () => {
		const users = [CURRENT_USER]
		const modelValue = {'🎉': users}
		const wrapper = mountReactions(modelValue)

		await wrapper.findAll('button')[0].trigger('click')
		await flushPromises()

		const emitted = wrapper.emitted('update:modelValue')
		expect(emitted).toHaveLength(1)
		expect(emitted![0][0]).toEqual({})
		expect(modelValue['🎉']).toBe(users)
		expect(users).toEqual([CURRENT_USER])
	})
})
