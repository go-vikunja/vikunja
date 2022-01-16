<template>
	<x-button
		variant="secondary"
		:icon="icon"
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:disabled="disabled || null"
		v-if="isButton"
	>
		{{ buttonText }}
	</x-button>
	<a
		v-tooltip="tooltipText"
		@click="changeSubscription"
		:class="{'is-disabled': disabled}"
		v-else
	>
		<span class="icon">
			<icon :icon="icon"/>
		</span>
		{{ buttonText }}
	</a>
</template>

<script lang="ts" setup>
import {computed, PropType, shallowRef} from 'vue'
import {useI18n} from 'vue-i18n'

import SubscriptionService from '@/services/subscription'
import SubscriptionModel from '@/models/subscription'

import {success} from '@/message'

const props = defineProps({
	entity: {
		required: true,
		type: String,
	},
	subscription: {
		required: true,
		type: Object as PropType<SubscriptionModel>,
	},
	entityId: {
		required: true,
		type: Number,
	},
	isButton: {
		type: Boolean,
		default: true,
	},
})

const subscriptionEntity = computed<string>(() => props.subscription.entity)

const emit = defineEmits(['change'])

const subscriptionService = shallowRef(new SubscriptionService())

const {t} = useI18n()
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
const icon = computed(() => props.subscription !== null ? ['far', 'bell-slash'] : 'bell')
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
