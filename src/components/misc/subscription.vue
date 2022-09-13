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
		:class="{'is-disabled': disabled}"
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
		return t('task.subscription.subscribedThroughParent', {
			entity: props.entity,
			parent: subscriptionEntity.value,
		})
	}

	return props.modelValue !== null ?
		t('task.subscription.subscribed', {entity: props.entity}) :
		t('task.subscription.notSubscribed', {entity: props.entity})
})

const buttonText = computed(() => props.modelValue ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed(() => props.modelValue ? ['far', 'bell-slash'] : 'bell')
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
	success({message: t('task.subscription.subscribeSuccess', {entity: props.entity})})
}

async function unsubscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.value.delete(subscription)
	emit('update:modelValue', null)
	success({message: t('task.subscription.unsubscribeSuccess', {entity: props.entity})})
}
</script>
