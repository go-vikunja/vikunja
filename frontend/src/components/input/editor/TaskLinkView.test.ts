import {describe, it, expect, beforeEach, vi} from 'vitest'
import {mount} from '@vue/test-utils'

const push = vi.fn()
vi.mock('vue-router', async (importOriginal) => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => ({push}),
	useRoute: () => ({fullPath: '/projects/1/2'}),
}))

import TaskLinkView from './TaskLinkView.vue'

const HREF = 'http://localhost:3000/tasks/5'

function mountView(isEditable: boolean) {
	return mount(TaskLinkView, {
		props: {
			node: {attrs: {href: HREF}},
			editor: {isEditable},
			decorations: [],
			selected: false,
			extension: {},
			getPos: () => 0,
			updateAttributes: () => {},
			deleteNode: () => {},
			view: {},
			innerDecorations: {},
			HTMLAttributes: {},
		} as never,
		global: {
			stubs: {
				NodeViewWrapper: {template: '<span><slot /></span>'},
				TaskLinkPill: {
					template: '<a @click="$emit(\'open\', {id: 5})" />',
					emits: ['open'],
				},
			},
		},
	})
}

describe('TaskLinkView', () => {
	beforeEach(() => {
		push.mockReset()
	})

	it('opens the task as a modal over the current route when not editable', async () => {
		const wrapper = mountView(false)

		await wrapper.find('a').trigger('click')

		expect(push).toHaveBeenCalledWith({
			name: 'task.detail',
			params: {id: 5},
			state: {backdropView: '/projects/1/2'},
		})
	})

	it('does not navigate while the editor is editable', async () => {
		const wrapper = mountView(true)

		await wrapper.find('a').trigger('click')

		expect(push).not.toHaveBeenCalled()
	})
})
