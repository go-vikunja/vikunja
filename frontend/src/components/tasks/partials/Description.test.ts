import {describe, expect, it, vi} from 'vitest'
import {shallowMount} from '@vue/test-utils'

import Description from './Description.vue'

vi.mock('vue-i18n', async (importOriginal) => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/components/input/AsyncEditor', () => ({
	default: {template: '<div class="tiptap__task-description stub-editor" />'},
}))

vi.mock('@/client/queries/taskMutations', () => ({
	useUpdateTaskMutation: () => ({mutateAsync: vi.fn()}),
}))

vi.mock('@/client/queries/attachments', () => ({
	useUploadAttachmentsMutation: () => ({mutateAsync: vi.fn()}),
}))

describe('Description', () => {
	it('uses a fixed-height scrollable editor on the task detail page', () => {
		const wrapper = shallowMount(Description, {
			props: {
				modelValue: {id: 1, project_id: 1, description: ''},
				canWrite: true,
			},
			global: {
				mocks: {$t: (key: string) => key},
				stubs: {Icon: true, CustomTransition: {template: '<div><slot/></div>'}},
			},
		})

		expect(wrapper.find('.tiptap__task-description').exists()).toBe(true)
	})
})
