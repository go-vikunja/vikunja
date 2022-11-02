<template>
	<x-button
		v-if="type === 'button'"
		variant="secondary"
		:icon="iconName"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled"
	>
		{{ buttonText }}
	</x-button>
	<DropdownItem
		v-else-if="type === 'dropdown'"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled"
		:icon="iconName"
	>
		{{ buttonText }}
	</DropdownItem>
	<BaseButton
		v-else
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:class="{'is-disabled': disabled}"
		:disabled="disabled"
	>
		<span class="icon">
			<icon :icon="iconName"/>
		</span>
		{{ buttonText }}
	</BaseButton>
</template>

<script lang="ts" setup>
import {computed, shallowRef, type PropType} from 'vue'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import DropdownItem from '@/components/misc/dropdown-item.vue'

import SubscriptionService from '@/services/subscription'
import SubscriptionModel from '@/models/subscription'
import type {ISubscription} from '@/modelTypes/ISubscription'

import {success} from '@/message'
import type { IconProp } from '@fortawesome/fontawesome-svg-core'

const props = defineProps({
	entity: String,
	entityId: Number,
	isButton: {
		type: Boolean,
		default: true,
	},
	modelValue: {
		type: Object as PropType<ISubscription>,
		default: null,
	},
	type: {
		type: String as PropType<'button' | 'dropdown' | 'null'>,
		default: 'button',
	},
})

const subscriptionEntity = computed<string | null>(() => props.modelValue?.entity ?? null)

const emit = defineEmits(['update:modelValue'])

const subscriptionService = shallowRef(new SubscriptionService())

const {t} = useI18n({useScope: 'global'})

const tooltipText = computed(() => {
	if (disabled.value) {
		if (props.entity === 'list' && subscriptionEntity.value === 'namespace') {
			return t('task.subscription.subscribedListThroughParentNamespace')
		}
		if (props.entity === 'task' && subscriptionEntity.value === 'namespace') {
			return t('task.subscription.subscribedTaskThroughParentNamespace')
		}
		if (props.entity === 'task' && subscriptionEntity.value === 'list') {
			return t('task.subscription.subscribedTaskThroughParentList')
		}

		return ''
	}

	switch (props.entity) {
		case 'namespace':
			return props.modelValue !== null ?
				t('task.subscription.subscribedNamespace') :
				t('task.subscription.notSubscribedNamespace')
		case 'list':
			return props.modelValue !== null ?
				t('task.subscription.subscribedList') :
				t('task.subscription.notSubscribedList')
		case 'task':
			return props.modelValue !== null ?
				t('task.subscription.subscribedTask') :
				t('task.subscription.notSubscribedTask')
	}

	return ''
})

const buttonText = computed(() => props.modelValue ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed<IconProp>(() => props.modelValue ? ['far', 'bell-slash'] : 'bell')
const disabled = computed(() => props.modelValue && subscriptionEntity.value !== props.entity)

function changeSubscription() {
	if (disabled.value) {
		return
	}

	if (props.modelValue === null) {
		subscribe()
	} else {
		unsubscribe()
	}
}

async function subscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.value.create(subscription)
	emit('update:modelValue', subscription)

	let message = ''
	switch (props.entity) {
		case 'namespace':
			message = t('task.subscription.subscribeSuccessNamespace')
			break
		case 'list':
			message = t('task.subscription.subscribeSuccessList')
			break
		case 'task':
			message = t('task.subscription.subscribeSuccessTask')
			break
	}
	success({message})
}

async function unsubscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.value.delete(subscription)
	emit('update:modelValue', null)

	let message = ''
	switch (props.entity) {
		case 'namespace':
			message = t('task.subscription.unsubscribeSuccessNamespace')
			break
		case 'list':
			message = t('task.subscription.unsubscribeSuccessList')
			break
		case 'task':
			message = t('task.subscription.unsubscribeSuccessTask')
			break
	}
	success({message})
}
</script>
