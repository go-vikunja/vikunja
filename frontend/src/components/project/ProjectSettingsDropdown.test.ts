import {shallowMount} from '@vue/test-utils'
import {defineComponent, h, nextTick} from 'vue'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const projectNavigation = vi.hoisted(() => ({
	invalidateProjects: vi.fn(() => Promise.resolve()),
	setProject: vi.fn(),
}))

vi.mock('@/composables/useProjectNavigation', () => ({
	useProjectNavigation: () => projectNavigation,
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({settings: {defaultProjectId: 0}}),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({enabledBackgroundProviders: []}),
}))

const DropdownStub = defineComponent({
	name: 'Dropdown',
	setup(_, {slots}) {
		return () => h('div', [
			slots.trigger?.({open: false, toggleOpen: () => {}}),
			slots.default?.(),
		])
	},
})

const SubscriptionStub = defineComponent({
	name: 'Subscription',
	props: {
		modelValue: {type: Object, default: null},
	},
	emits: ['update:modelValue'],
	template: '<div />',
})

import ProjectSettingsDropdown from './ProjectSettingsDropdown.vue'
import type {ISubscription} from '@/modelTypes/ISubscription'
import UserModel from '@/models/user'

describe('ProjectSettingsDropdown subscriptions', () => {
	beforeEach(() => {
		projectNavigation.invalidateProjects.mockClear()
		projectNavigation.setProject.mockClear()
	})

	it('updates local state and invalidates project caches after a subscription change', async () => {
		const wrapper = shallowMount(ProjectSettingsDropdown, {
			props: {
				project: {id: 1, title: 'Project'},
			},
			global: {
				stubs: {
					BaseButton: true,
					Dropdown: DropdownStub,
					DropdownItem: true,
					Icon: true,
					Subscription: SubscriptionStub,
				},
				directives: {tooltip: () => {}},
				mocks: {$t: (key: string) => key},
			},
		})
		const subscription: ISubscription = {
			id: 7,
			entity: 'project',
			entityId: 1,
			user: new UserModel({id: 1, username: 'test', name: 'Test User'}),
			created: new Date('2026-09-01T00:00:00Z'),
			maxPermission: null,
		}
		const subscriptionComponent = wrapper.findComponent(SubscriptionStub)

		subscriptionComponent.vm.$emit('update:modelValue', subscription)
		await nextTick()

		expect(subscriptionComponent.props('modelValue')).toEqual(subscription)
		expect(projectNavigation.invalidateProjects).toHaveBeenCalledOnce()
		expect(projectNavigation.setProject).not.toHaveBeenCalled()
	})
})
