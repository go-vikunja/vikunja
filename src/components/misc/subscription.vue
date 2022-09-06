<template>
	<x-button
		v-if="type === 'button'"
		variant="secondary"
		:icon="iconName"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled || undefined"
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
	subscription: {
		type: Object as PropType<ISubscription>,
		default: null,
	},
	type: {
		type: String as PropType<'button' | 'dropdown' | 'null'>,
		default: 'button',
	},
})

const subscriptionEntity = computed<string | null>(() => props.subscription?.entity ?? null)

const emit = defineEmits(['change'])

const subscriptionService = shallowRef(new SubscriptionService())

const {t} = useI18n({useScope: 'global'})
const tooltipText = computed(() => {
	if (disabled.value) {
		return t('task.subscription.subscribedThroughParent', {
			entity: props.entity,
			parent: subscriptionEntity.value,
		})
	}

	return props.subscription !== null ?
		t('task.subscription.subscribed', {entity: props.entity}) :
		t('task.subscription.notSubscribed', {entity: props.entity})
})

const buttonText = computed(() => props.subscription !== null ? t('task.subscription.unsubscribe') : t('task.subscription.subscribe'))
const iconName = computed(() => props.subscription !== null ? ['far', 'bell-slash'] : 'bell')
const disabled = computed(() => {
	if (props.subscription === null) {
		return false
	}

	return subscriptionEntity.value !== props.entity
})

function changeSubscription() {
	if (disabled.value) {
		return
	}

	if (props.subscription === null) {
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
	emit('change', subscription)
	success({message: t('task.subscription.subscribeSuccess', {entity: props.entity})})
}

async function unsubscribe() {
	const subscription = new SubscriptionModel({
		entity: props.entity,
		entityId: props.entityId,
	})
	await subscriptionService.value.delete(subscription)
	emit('change', null)
	success({message: t('task.subscription.unsubscribeSuccess', {entity: props.entity})})
}
</script>
