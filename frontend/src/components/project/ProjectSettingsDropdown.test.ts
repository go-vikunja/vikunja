import {shallowMount} from '@vue/test-utils'
import {defineComponent, h} from 'vue'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const subscriptionMutation = vi.hoisted(() => ({
	mutateAsync: vi.fn(() => Promise.resolve(null)),
}))

vi.mock('@/client/queries/projects', async importOriginal => ({
	...await importOriginal<typeof import('@/client/queries/projects')>(),
	useSetProjectSubscriptionMutation: () => subscriptionMutation,
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({settings: {default_project_id: 0}}),
}))

vi.mock('@/stores/config', () => ({
	useConfigStore: () => ({enabled_background_providers: []}),
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
		modelValue: {type: Object},
	},
	emits: ['toggle'],
	template: '<div />',
})

import ProjectSettingsDropdown from './ProjectSettingsDropdown.vue'

function mountDropdown(project: InstanceType<typeof ProjectSettingsDropdown>['$props']['project']) {
	return shallowMount(ProjectSettingsDropdown, {
		props: {project},
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
}

describe('ProjectSettingsDropdown subscriptions', () => {
	beforeEach(() => {
		subscriptionMutation.mutateAsync.mockClear()
	})

	it('passes the project subscription to the subscription button', () => {
		const wrapper = mountDropdown({
			id: 1,
			title: 'Project',
			subscription: {id: 7, entity: 'project', entity_id: 1},
		})

		expect(wrapper.findComponent(SubscriptionStub).props('modelValue')).toMatchObject({
			id: 7,
			entity: 'project',
			entity_id: 1,
		})
	})

	it('passes null for a project without a subscription', () => {
		const wrapper = mountDropdown({id: 1, title: 'Project'})

		expect(wrapper.findComponent(SubscriptionStub).props('modelValue')).toBeNull()
	})

	it('toggles the subscription through the project mutation', () => {
		const wrapper = mountDropdown({id: 1, title: 'Project'})

		wrapper.findComponent(SubscriptionStub).vm.$emit('toggle', true)

		expect(subscriptionMutation.mutateAsync).toHaveBeenCalledWith({projectId: 1, subscribed: true})
	})
})
