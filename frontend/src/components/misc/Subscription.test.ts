import {mount} from '@vue/test-utils'
import {
	expect,
	it,
} from 'vitest'
import {createI18n} from 'vue-i18n'

import Subscription from './Subscription.vue'
import type {Subscription as SubscriptionModel} from '@/client/generated'
import en from '@/i18n/lang/en.json'

// vue-i18n compiles `en` in place, so expected text can't be read back from it.
const SUBSCRIBED_THROUGH_PROJECT = 'You are subscribed to this task through its project. Unsubscribing here only stops notifications for this task.'
const SUBSCRIBED_TASK = 'You are currently subscribed to this task and will receive notifications for changes.'
const NOT_SUBSCRIBED_TASK = 'You are not subscribed to this task and won\'t receive notifications for changes.'
const SUBSCRIBED_THROUGH_PARENT_PROJECT = 'You are subscribed to this project through a parent project. Unsubscribing here stops notifications for this project and everything in it.'
const SUBSCRIBED_PROJECT = 'You are currently subscribed to this project and will receive notifications for changes.'

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

function mountSubscription({
	modelValue,
	entity = 'task',
	entityId = 5,
}: {
	modelValue?: SubscriptionModel | null,
	entity?: NonNullable<SubscriptionModel['entity']>,
	entityId?: number,
}) {
	const tooltips: unknown[] = []
	mount(Subscription, {
		props: {
			modelValue,
			entity,
			entityId,
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
	expect(mountSubscription({modelValue: {
		id: 1,
		entity: 'project',
		entity_id: 2,
	}})).toEqual([SUBSCRIBED_THROUGH_PROJECT])
})

it('treats a subscription on the same entity as direct', () => {
	expect(mountSubscription({modelValue: {
		id: 1,
		entity: 'task',
		entity_id: 5,
	}})).toEqual([SUBSCRIBED_TASK])
})

it('marks a project subscription inherited from a parent project', () => {
	expect(mountSubscription({
		modelValue: {
			id: 1,
			entity: 'project',
			entity_id: 2,
		},
		entity: 'project',
		entityId: 7,
	})).toEqual([SUBSCRIBED_THROUGH_PARENT_PROJECT])
})

it('treats a subscription on the same project as direct', () => {
	expect(mountSubscription({
		modelValue: {
			id: 1,
			entity: 'project',
			entity_id: 7,
		},
		entity: 'project',
		entityId: 7,
	})).toEqual([SUBSCRIBED_PROJECT])
})

it('treats an undefined subscription as not subscribed', () => {
	expect(mountSubscription({modelValue: undefined})).toEqual([NOT_SUBSCRIBED_TASK])
})

it('treats a null subscription as not subscribed', () => {
	expect(mountSubscription({modelValue: null})).toEqual([NOT_SUBSCRIBED_TASK])
})
