import {mount} from '@vue/test-utils'
import {
	expect,
	it,
} from 'vitest'
import {createI18n} from 'vue-i18n'

import Subscription from './Subscription.vue'
import en from '@/i18n/lang/en.json'

// vue-i18n compiles `en` in place, so expected text can't be read back from it.
const SUBSCRIBED_THROUGH_PROJECT = 'You are subscribed to this task through its project. Unsubscribing here only stops notifications for this task.'
const SUBSCRIBED_TASK = 'You are currently subscribed to this task and will receive notifications for changes.'
const NOT_SUBSCRIBED_TASK = 'You are not subscribed to this task and won\'t receive notifications for changes.'

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

function mountSubscription(modelValue?: {
	id: number,
	entity: 'project' | 'task',
	entity_id: number,
}) {
	const tooltips: unknown[] = []
	mount(Subscription, {
		props: {
			modelValue,
			entity: 'task',
			entityId: 5,
		},
		global: {
			plugins: [i18n],
			stubs: {XButton: {template: '<button><slot /></button>'}},
			directives: {
				tooltip: {
					mounted: (_el, binding) => {
						tooltips.push(binding.value)
					},
				},
			},
		},
	})
	return tooltips
}

it('marks a task subscription inherited from another entity', () => {
	expect(mountSubscription({
		id: 1,
		entity: 'project',
		entity_id: 2,
	})).toEqual([SUBSCRIBED_THROUGH_PROJECT])
})

it('treats a subscription on the same entity as direct', () => {
	expect(mountSubscription({
		id: 1,
		entity: 'task',
		entity_id: 5,
	})).toEqual([SUBSCRIBED_TASK])
})

it('treats an undefined subscription as not subscribed', () => {
	expect(mountSubscription(undefined)).toEqual([NOT_SUBSCRIBED_TASK])
})
