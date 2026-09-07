import {shallowMount} from '@vue/test-utils'
import {nextTick} from 'vue'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'
import {i18n as globalI18n} from '@/i18n'
import {useAuthStore} from '@/stores/auth'

const labelState = vi.hoisted(() => ({
	labels: undefined as {value: Array<{id?: number, title?: string, hex_color?: string}>} | undefined,
	isPending: undefined as {value: boolean} | undefined,
}))

const editorState = vi.hoisted(() => ({
	text: '',
	options: undefined as {onUpdate: (context: {editor: unknown}) => void} | undefined,
	editor: undefined as Record<string, unknown> | undefined,
	dispatch: vi.fn(),
}))

vi.mock('@/composables/useLabels', async () => {
	const {ref} = await import('vue')
	labelState.labels = ref([])
	labelState.isPending = ref(false)

	return {
		useLabels: () => ({
			labels: labelState.labels,
			isPending: labelState.isPending,
			getLabelById: () => null,
			getLabelByExactTitle: () => null,
		}),
	}
})

vi.mock('@/stores/projects', () => ({
	useProjectStore: () => ({
		projects: {},
		findProjectByExactname: () => null,
	}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@tiptap/vue-3', async () => {
	const {defineComponent, h, shallowRef} = await import('vue')
	const editor = {
		getText: () => editorState.text,
		commands: {
			setContent: (content: string | {content?: Array<{content?: Array<{text?: string}>}>}) => {
				editorState.text = typeof content === 'string'
					? content
					: content.content?.[0]?.content?.[0]?.text ?? ''
			},
			setTextSelection: vi.fn(),
			focus: vi.fn(),
		},
		state: {
			selection: {from: 1},
			doc: {content: {size: 100}},
			tr: {setMeta: vi.fn().mockReturnThis()},
		},
		view: {dispatch: editorState.dispatch},
		destroy: vi.fn(),
	}
	editorState.editor = editor

	return {
		EditorContent: defineComponent({render: () => h('div')}),
		useEditor: (options: typeof editorState.options) => {
			editorState.options = options
			return shallowRef(editor)
		},
		VueRenderer: class {},
	}
})

import FilterInput from './FilterInput.vue'

const JALALI_DIGITS = new RegExp('[۰-۹٠-٩]')

function mountInput(modelValue = '') {
	return shallowMount(FilterInput, {
		props: {modelValue},
		global: {mocks: {$t: (key: string) => key}},
	})
}

function emitThroughEditor(text: string) {
	editorState.text = text
	editorState.options!.onUpdate({editor: editorState.editor!})
}

describe('FilterInput Jalali typed dates', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		editorState.text = ''
		globalI18n.global.locale.value = 'fa-IR' as never
		useAuthStore().setUserSettings({timezone: 'Asia/Tehran'} as never)
	})

	afterEach(() => {
		globalI18n.global.locale.value = 'en' as never
		vi.useRealTimers()
	})

	it.each([
		['Asia/Tehran', '2026-09-06T20:30:00.000Z'],
		['UTC', '2026-09-07T00:00:00.000Z'],
		['America/Los_Angeles', '2026-09-07T07:00:00.000Z'],
	])('converts typed 1405/06/16 in %s to Gregorian ISO', async (timezone, expected) => {
		useAuthStore().setUserSettings({timezone} as never)
		const wrapper = mountInput()
		emitThroughEditor('dueDate > 1405/06/16')
		await nextTick()
		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as string
		expect(emitted).toBe(`due_date > ${expected}`)
		expect(emitted).not.toMatch(JALALI_DIGITS)
		wrapper.unmount()
	})

	it('converts Persian digits with time', async () => {
		const wrapper = mountInput()
		emitThroughEditor('dueDate > ۱۴۰۵/۰۶/۱۶ ۱۵:۳۰')
		await nextTick()
		const emitted = wrapper.emitted('update:modelValue')?.pop()?.[0] as string
		expect(emitted).toBe('due_date > 2026-09-07T12:00:00.000Z')
		expect(emitted).not.toMatch(JALALI_DIGITS)
		wrapper.unmount()
	})

	it('keeps datemath and ISO verbatim (fall back to existing path)', async () => {
		const wrapper = mountInput()
		emitThroughEditor('dueDate = now/d')
		await nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date = now/d')
		emitThroughEditor('dueDate > 2026-09-07')
		await nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date > 2026-09-07')
		wrapper.unmount()
	})

	it('keeps day presets verbatim in fa mode', async () => {
		const wrapper = mountInput()
		emitThroughEditor('dueDate > now+7d')
		await nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date > now+7d')
		wrapper.unmount()
	})

	it('keeps typed week datemath verbatim so saved filters stay dynamic', async () => {
		vi.useFakeTimers()
		vi.setSystemTime(new Date('2026-09-07T12:00:00.000Z'))
		const wrapper = mountInput()
		emitThroughEditor('dueDate > now/w')
		await nextTick()
		// Verbatim: evaluated dynamically (documented Monday-start) rather than
		// frozen at edit time. UI preset picks still resolve to explicit bounds.
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date > now/w')
		wrapper.unmount()
	})

	it.each([['en'], ['de']])('leaves Jalali-looking text untouched in %s (byte-identical)', async (locale) => {
		globalI18n.global.locale.value = locale as never
		const wrapper = mountInput()
		emitThroughEditor('dueDate > 1405/06/16')
		await nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date > 1405/06/16')
		emitThroughEditor('dueDate > now/w')
		await nextTick()
		expect(wrapper.emitted('update:modelValue')?.pop()?.[0]).toBe('due_date > now/w')
		wrapper.unmount()
	})
})
