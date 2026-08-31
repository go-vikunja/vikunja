import {shallowMount} from '@vue/test-utils'
import {nextTick} from 'vue'
import {beforeEach, describe, expect, it, vi} from 'vitest'

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
	labelState.isPending = ref(true)

	return {
		useLabels: () => ({
			labels: labelState.labels,
			isPending: labelState.isPending,
			getLabelById: (id: number) => labelState.labels?.value.find(label => label.id === id),
			getLabelByExactTitle: (title: string) => labelState.labels?.value.find(label => label.title === title),
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

describe('FilterInput label loading', () => {
	beforeEach(() => {
		editorState.text = ''
		editorState.dispatch.mockReset()
		labelState.labels!.value = []
		labelState.isPending!.value = true
	})

	it('resolves label ids and refreshes highlighting when labels finish loading', async () => {
		const wrapper = shallowMount(FilterInput, {
			props: {modelValue: 'labels = 1'},
			global: {mocks: {$t: (key: string) => key}},
		})
		expect(editorState.text).toBe('labels = 1')

		labelState.labels!.value = [{id: 1, title: 'Work', hex_color: 'ff006e'}]
		labelState.isPending!.value = false
		await nextTick()

		expect(editorState.text).toBe('labels = Work')
		expect(editorState.dispatch).toHaveBeenCalled()
		wrapper.unmount()
	})

	it('does not replace an edit made before labels finish loading', async () => {
		const wrapper = shallowMount(FilterInput, {
			props: {modelValue: 'labels = 1'},
			global: {mocks: {$t: (key: string) => key}},
		})
		editorState.text = 'labels = Personal'
		editorState.options!.onUpdate({editor: editorState.editor!})

		labelState.labels!.value = [{id: 1, title: 'Work'}]
		labelState.isPending!.value = false
		await nextTick()

		expect(editorState.text).toBe('labels = Personal')
		wrapper.unmount()
	})
})
