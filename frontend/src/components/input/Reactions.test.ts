import {describe, it, expect, vi} from 'vitest'
import {nextTick} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import {reactionsCreate} from '@/client/generated'
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
		settings: {frontend_settings: {color_schema: 'light'}},
	}),
}))

vi.mock('@/client/generated', () => ({
	reactionsCreate: vi.fn(async () => ({data: {}})),
	reactionsDelete: vi.fn(async () => ({data: undefined})),
}))
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

vi.mock('vuemoji-picker', () => ({
	VuemojiPicker: {
		name: 'VuemojiPicker',
		template: '<div />',
		emits: ['emojiClick'],
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountReactions(
	modelValue: Record<string, typeof CURRENT_USER[]>,
	disabled = false,
	entityKind: 'tasks' | 'comments' = 'tasks',
) {
	const queryClient = new QueryClient()
	const subjectProps = entityKind === 'comments'
		? {entityKind, taskId: 1}
		: {entityKind}
	return mount(Reactions, {
		props: {
			...subjectProps,
			entityId: 1,
			modelValue,
			disabled,
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

function deferReactionCreate() {
	let resolveCreate: (value: unknown) => void = () => {}
	const pending = new Promise<unknown>(resolve => {
		resolveCreate = resolve
	}) as ReturnType<typeof reactionsCreate>
	vi.mocked(reactionsCreate).mockReturnValueOnce(pending)
	return () => resolveCreate({data: {}})
}

describe('Reactions', () => {
	it('marks the reacted-by-current-user button as pressed', () => {
		const wrapper = mountReactions({
			'🎉': [CURRENT_USER],
			'👍': [OTHER_USER],
		})

		expect(wrapper.findAll('button')[0].attributes('aria-pressed')).toBe('true')
		expect(wrapper.findAll('button')[1].attributes('aria-pressed')).toBe('false')
	})

	it('disables reaction buttons natively when the viewer may not write', () => {
		const wrapper = mountReactions({'🎉': [OTHER_USER]}, true)

		const button = wrapper.findAll('button')[0]
		expect(button.attributes('disabled')).toBeDefined()
		expect(button.attributes('aria-disabled')).toBeUndefined()
	})

	it('keeps reaction buttons focusable while the mutation is pending', async () => {
		const resolveCreate = deferReactionCreate()
		const wrapper = mountReactions({'🎉': [OTHER_USER]})

		await wrapper.findAll('button')[0].trigger('click')
		await nextTick()

		expect(wrapper.findAll('button')[0].attributes('aria-disabled')).toBe('true')
		expect(wrapper.findAll('button')[0].attributes('disabled')).toBeUndefined()

		resolveCreate()
		await flushPromises()

		expect(wrapper.findAll('button')[0].attributes('aria-disabled')).toBeUndefined()
	})

	it('keeps the emoji picker open when the response resolves after the entity changed', async () => {
		const resolveCreate = deferReactionCreate()
		const wrapper = mountReactions({'🎉': [OTHER_USER]}, false, 'comments')

		await wrapper.findAll('button')[1].trigger('click')
		await nextTick()
		expect(wrapper.find('.emoji-picker').exists()).toBe(true)

		wrapper.findComponent({name: 'VuemojiPicker'}).vm.$emit('emojiClick', {unicode: '👍'})
		await nextTick()
		await wrapper.setProps({entityId: 2})

		resolveCreate()
		await flushPromises()

		expect(wrapper.find('.emoji-picker').exists()).toBe(true)
	})
})
